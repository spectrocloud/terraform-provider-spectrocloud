#!/usr/bin/env python3
"""
Spectro Terraform Formatter
Automatically formats generated Terraform files by extracting YAML content
and creating templated configuration files.

Usage: spectro-tf-format [options] <terraform_file>
"""

import subprocess
import sys
import argparse
from pathlib import Path
import yaml
import os
import re
import shutil
import platform
from typing import Dict, List, Any, Tuple, Optional

# Platform detection for cross-platform compatibility
IS_WINDOWS = platform.system() == 'Windows'
IS_POSIX = os.name == 'posix'


# Built-in templating configuration for Spectro Cloud
BUILTIN_TEMPLATING_CONFIG = {
    # Script configuration defaults
    "rules": ["yaml-format"],
    "output_dir": "cluster_configs_yaml",
    "backup": True,
    
    # YAML processing configuration
    "cloud_config.values": {
        "filename": "{resource_name}_cloud_config.yaml",
        "templating": [
            # {"path": "Cluster.spec.infrastructureRef.name", "variable": "CLUSTER_NAME"},
            # {"path": "AWSCluster.metadata.name", "variable": "CLUSTER_NAME"},
        ]
    },
    "machine_pool.node_pool_config": {
        "filename": "{resource_name}_{pool_name}_config.yaml",
        "templating": [
            # Control plane paths
            {"path": "KubeadmControlPlane.spec.replicas", "variable": "REPLICAS"},
            {"path": "KubeadmControlPlane.spec.machineTemplate.infrastructureRef.name", "variable": "MACHINE_TEMPLATE_NAME"},
            
            # Worker pool paths (alternative paths for MachineDeployment)
            {"path": "MachineDeployment.spec.template.spec.infrastructureRef.name", "variable": "MACHINE_TEMPLATE_NAME"},
            {"path": "MachineDeployment.spec.template.spec.bootstrap.configRef.name", "variable": "KC_TEMPLATE_NAME"},
            {"path": "KubeadmConfigTemplate.metadata.name", "variable": "KC_TEMPLATE_NAME"},

            # {"path": "MachineDeployment.metadata.name", "variable": "POOL_NAME"},

            # Common paths (works for both)
            {"path": "AWSMachineTemplate.metadata.name", "variable": "MACHINE_TEMPLATE_NAME"},
            {"path": "MachineDeployment.spec.replicas", "variable": "REPLICAS"},
            {"path": "AWSMachineTemplate.spec.template.spec.ami.id", "variable": "AMI_ID"},
            {"path": "AWSMachineTemplate.spec.template.spec.instanceType", "variable": "INSTANCE_TYPE"},
            {"path": "AWSMachineTemplate.spec.template.spec.rootVolume.size", "variable": "ROOT_VOLUME_SIZE"}
        ]
    }
}


class YAMLTemplater:
    """Handles YAML templating and variable extraction"""
    
    def process_yaml(self, yaml_content: str, templating_config: List[Dict[str, str]]) -> Tuple[str, Dict[str, Any]]:
        """Process YAML content and extract variables according to templating config."""
        documents, original_raw_documents = self._split_yaml_documents_with_raw(yaml_content)
        extracted_vars = {}
        
        if not documents:
            print(f"    ⚠️  No valid YAML documents found - returning original content")
            return yaml_content, extracted_vars
        
        print(f"    📄 Processing {len(documents)} parsed documents + {len(original_raw_documents) - len(documents)} raw documents")
        
        # Log document types for debugging
        doc_types = []
        for i, doc in enumerate(documents):
            if isinstance(doc, dict) and 'kind' in doc:
                doc_types.append(f"{doc['kind']}")
            else:
                doc_types.append("unknown")
        print(f"    📋 Document types found: {', '.join(doc_types)}")
        
        # Apply conditional logic for worker pools
        enhanced_templating_config = self._apply_conditional_worker_logic(documents, templating_config)
        
        for template_config in enhanced_templating_config:
            doc_type, actual_path = self._parse_document_path(template_config["path"])
            found = False
            
            for i, doc in enumerate(documents):
                # If document type is specified, only process matching documents
                if doc_type:
                    if not isinstance(doc, dict) or doc.get('kind') != doc_type:
                        continue
                        
                try:
                    value = self._extract_value_from_path(doc, actual_path)
                    if value is not None:
                        extracted_vars[template_config["variable"]] = value
                        self._replace_value_in_doc(doc, actual_path, f"${{{template_config['variable']}}}")
                        print(f"    ✓ Found {template_config['variable']} = {value} at path {template_config['path']}")
                        found = True
                        break  # Stop after first match
                except Exception as e:
                    continue
            
            if not found:
                print(f"    ✗ Path {template_config['path']} not found in any document")
        
        # Reconstruct YAML preserving original structure
        try:
            modified_content = self._reconstruct_yaml_with_fallback(documents, original_raw_documents, extracted_vars)
        except Exception as e:
            print(f"    ⚠️  Error reconstructing YAML: {e} - returning original content")
            return yaml_content, extracted_vars
            
        return modified_content, extracted_vars

    def _apply_conditional_worker_logic(self, documents: List[Dict], templating_config: List[Dict[str, str]]) -> List[Dict[str, str]]:
        """Apply conditional logic for worker pools based on name matching."""
        enhanced_config = templating_config.copy()
        
        # Find MachineDeployment and AWSMachineTemplate documents
        machine_deployment_doc = None
        aws_machine_template_doc = None
        
        for doc in documents:
            if isinstance(doc, dict) and doc.get('kind') == 'MachineDeployment':
                machine_deployment_doc = doc
            elif isinstance(doc, dict) and doc.get('kind') == 'AWSMachineTemplate':
                aws_machine_template_doc = doc
        
        # Only apply logic if both documents exist (indicating this is a worker pool)
        if machine_deployment_doc and aws_machine_template_doc:
            # Extract metadata.name from both documents
            md_name = self._extract_value_from_path(machine_deployment_doc, "metadata.name")
            aws_template_name = self._extract_value_from_path(aws_machine_template_doc, "metadata.name")
            
            print(f"    🔍 Worker pool detected: MachineDeployment.name='{md_name}', AWSMachineTemplate.name='{aws_template_name}'")
            
            # Find the variable used for AWSMachineTemplate.metadata.name
            aws_template_variable = None
            for config in templating_config:
                if config.get("path") == "AWSMachineTemplate.metadata.name":
                    aws_template_variable = config.get("variable")
                    break
            
            if md_name and aws_template_name and aws_template_variable:
                if md_name == aws_template_name:
                    # Names match - add conditional rule for MachineDeployment.metadata.name
                    conditional_rule = {
                        "path": "MachineDeployment.metadata.name", 
                        "variable": aws_template_variable
                    }
                    enhanced_config.append(conditional_rule)
                    print(f"    ✅ Names match! Added rule: MachineDeployment.metadata.name → {aws_template_variable}")
                else:
                    # Names don't match - don't template MachineDeployment.metadata.name
                    print(f"    ❌ Names don't match. MachineDeployment.metadata.name will NOT be templated")
        
        return enhanced_config

    def _split_yaml_documents_with_raw(self, yaml_content: str) -> Tuple[List[Dict], List[str]]:
        """Split YAML content into parsed documents and original raw document strings."""
        # First split into raw document strings
        raw_documents = yaml_content.split('---')
        raw_documents = [doc.strip() for doc in raw_documents if doc.strip()]
        
        # Try to parse each document
        parsed_documents = []
        
        for i, raw_doc in enumerate(raw_documents):
            try:
                parsed_doc = yaml.safe_load(raw_doc)
                if parsed_doc is not None and isinstance(parsed_doc, dict):
                    parsed_documents.append(parsed_doc)
                    print(f"    ✓ Successfully parsed raw document {i+1}")
                else:
                    print(f"    ✗ Raw document {i+1} parsed but is not a valid dict")
            except yaml.YAMLError as e:
                print(f"    ✗ Could not parse raw document {i+1}: {e}")
                # We'll keep this as a raw document in the fallback reconstruction
                continue
            except Exception as e:
                print(f"    ✗ Unexpected error parsing raw document {i+1}: {e}")
                continue
        
        return parsed_documents, raw_documents

    def _reconstruct_yaml_with_fallback(self, parsed_docs: List[Dict], raw_docs: List[str], extracted_vars: Dict[str, Any]) -> str:
        """Reconstruct YAML preserving unparseable documents in original form."""
        result_parts = []
        
        # Track which raw documents were successfully parsed
        parsed_count = 0
        
        for i, raw_doc in enumerate(raw_docs):
            try:
                # Try to parse this raw document
                test_parsed = yaml.safe_load(raw_doc)
                if test_parsed is not None and isinstance(test_parsed, dict):
                    # This document was successfully parsed, use the processed version
                    if parsed_count < len(parsed_docs):
                        processed_yaml = yaml.dump(parsed_docs[parsed_count], default_flow_style=False, sort_keys=False)
                        result_parts.append(processed_yaml.strip())
                        parsed_count += 1
                    else:
                        # Fallback to raw if we somehow don't have a parsed version
                        result_parts.append(raw_doc)
                else:
                    # Document didn't parse to a valid dict, use raw
                    result_parts.append(raw_doc)
            except:
                # Document failed to parse, preserve it as-is but try to apply variable substitution manually
                print(f"    📝 Preserving unparseable document {i+1} with manual variable substitution")
                substituted_doc = self._apply_manual_variable_substitution(raw_doc, extracted_vars)
                result_parts.append(substituted_doc)
        
        # Join with document separators
        return '---\n' + '\n---\n'.join(result_parts)

    def _apply_manual_variable_substitution(self, raw_content: str, extracted_vars: Dict[str, Any]) -> str:
        """Apply variable substitution to raw YAML content that couldn't be parsed."""
        result = raw_content
        
        # First, fix multi-line string formatting issues that cause YAML parsing failures
        result = self._fix_multiline_yaml_strings(result)
        
        # Apply simple string replacements for known patterns
        # This is a basic approach for common cases
        for var_name, var_value in extracted_vars.items():
            # Look for specific patterns that might need replacement
            if var_name == 'MACHINE_TEMPLATE_NAME':
                # Replace machine template name references - only in specific contexts
                # Target infrastructureRef.name specifically (handles nested infrastructureRef)
                result = re.sub(r'(infrastructureRef:\s*\n(?:.*\n)*?\s*name:\s+)[^\s\n]+', 
                               f'\\1${{{var_name}}}', result)
                # Target AWSMachineTemplate metadata.name
                result = re.sub(r'(AWSMachineTemplate\s*\n.*?metadata:\s*\n.*?name:\s+)[^\s\n]+', 
                               f'\\1${{{var_name}}}', result, flags=re.DOTALL)
                # Target MachineDeployment metadata.name (conditional - only when names match)
                result = re.sub(r'(MachineDeployment\s*\n.*?metadata:\s*\n.*?name:\s+)[^\s\n]+', 
                               f'\\1${{{var_name}}}', result, flags=re.DOTALL)
            elif var_name == 'REPLICAS':
                # Replace replica counts
                result = re.sub(r'(replicas:\s+)\d+', f'\\1${{{var_name}}}', result)
        
        return result
    
    def _fix_multiline_yaml_strings(self, yaml_content: str) -> str:
        """Fix common multi-line string formatting issues in YAML."""
        lines = yaml_content.split('\n')
        fixed_lines = []
        i = 0
        
        while i < len(lines):
            line = lines[i]
            
            # Check for lines ending with backslash (shell line continuation)
            if line.rstrip().endswith('\\'):
                # Found a multi-line shell command with backslashes
                # Get the original indentation and YAML list marker
                stripped_line = line.lstrip()
                original_indent = len(line) - len(stripped_line)
                
                # Extract the command part (everything after the YAML list marker)
                if stripped_line.startswith('- '):
                    command_start = stripped_line[2:].rstrip()[:-1]  # Remove '- ' and trailing backslash
                else:
                    command_start = stripped_line.rstrip()[:-1]  # Just remove backslash
                
                # Collect all parts of the command
                command_parts = [command_start.strip()]
                i += 1
                
                # Collect continuation lines
                while i < len(lines):
                    current_line = lines[i]
                    if current_line.strip().endswith('\\'):
                        # This is a continuation line
                        command_parts.append(current_line.strip()[:-1])  # Remove backslash
                        i += 1
                    else:
                        # This is the final line of the command
                        command_parts.append(current_line.strip())
                        i += 1
                        break
                
                # Join the command parts with actual newlines
                # This preserves the original shell command structure  
                full_command = '\\n'.join(command_parts)
                
                # Escape double quotes within the command for proper YAML formatting
                escaped_command = full_command.replace('"', '\\"')
                
                # Reconstruct the YAML list item with proper indentation
                indent = ' ' * original_indent
                fixed_lines.append(f'{indent}- "{escaped_command}"')
            else:
                fixed_lines.append(line)
                i += 1
        
        return '\n'.join(fixed_lines)

    def _split_yaml_documents(self, yaml_content: str) -> List[Dict]:
        """Split multi-document YAML content into individual parsed documents."""
        try:
            # Use yaml.safe_load_all to properly parse multi-document YAML
            documents = list(yaml.safe_load_all(yaml_content))
            # Filter out None documents
            valid_docs = [doc for doc in documents if doc is not None and isinstance(doc, dict)]
            if not valid_docs:
                print(f"    Warning: No valid YAML documents found in content with standard parser")
                # Try individual parsing as fallback
                return self._parse_yaml_documents_individually(yaml_content)
            return valid_docs
        except yaml.YAMLError as e:
            print(f"    Warning: Could not parse YAML with standard parser: {e}")
            # Always try individual document parsing as fallback
            return self._parse_yaml_documents_individually(yaml_content)
    
    def _parse_yaml_documents_individually(self, yaml_content: str) -> List[Dict]:
        """Try to parse YAML documents individually with multiple strategies."""
        valid_docs = []
        
        # Split on document separators
        yaml_docs = yaml_content.split('---')
        
        for i, doc_content in enumerate(yaml_docs):
            doc_content = doc_content.strip()
            if not doc_content:
                continue
                
            parsed_doc = None
            parsing_method = "unknown"
            
            # Strategy 1: Regular YAML parsing
            try:
                parsed_doc = yaml.safe_load(doc_content)
                if parsed_doc is not None and isinstance(parsed_doc, dict):
                    parsing_method = "standard"
                else:
                    parsed_doc = None
            except yaml.YAMLError:
                parsed_doc = None
            
            # Strategy 2: Try with custom YAML loader that's more permissive
            if parsed_doc is None:
                try:
                    from yaml import SafeLoader
                    class PermissiveLoader(SafeLoader):
                        pass
                    
                    # Add custom handling for problematic constructs
                    def construct_undefined(loader, node):
                        if isinstance(node, yaml.ScalarNode):
                            return loader.construct_scalar(node)
                        elif isinstance(node, yaml.SequenceNode):
                            return loader.construct_sequence(node)
                        elif isinstance(node, yaml.MappingNode):
                            return loader.construct_mapping(node)
                        return None
                    
                    PermissiveLoader.add_constructor(None, construct_undefined)
                    
                    parsed_doc = yaml.load(doc_content, Loader=PermissiveLoader)
                    if parsed_doc is not None and isinstance(parsed_doc, dict):
                        parsing_method = "permissive"
                    else:
                        parsed_doc = None
                except Exception:
                    parsed_doc = None
            
            # Strategy 3: Try with line-by-line reconstruction for common issues
            if parsed_doc is None:
                try:
                    cleaned_content = self._clean_yaml_content(doc_content)
                    parsed_doc = yaml.safe_load(cleaned_content)
                    if parsed_doc is not None and isinstance(parsed_doc, dict):
                        parsing_method = "cleaned"
                    else:
                        parsed_doc = None
                except Exception:
                    parsed_doc = None
            
            if parsed_doc is not None:
                valid_docs.append(parsed_doc)
                print(f"    ✓ Successfully parsed YAML document {i+1} using {parsing_method} method")
            else:
                print(f"    ✗ Could not parse YAML document {i+1} with any method")
                # Show a snippet for debugging
                lines = doc_content.split('\n')
                if len(lines) > 10:
                    snippet = '\n'.join(lines[:5]) + '\n  ... (' + str(len(lines)-10) + ' more lines) ...\n' + '\n'.join(lines[-5:])
                else:
                    snippet = doc_content
                print(f"      Content snippet:\n{snippet}")
        
        print(f"    → Successfully parsed {len(valid_docs)} out of {len([d for d in yaml_docs if d.strip()])} documents")
        return valid_docs
    
    def _clean_yaml_content(self, yaml_content: str) -> str:
        """Clean YAML content to fix common structural issues."""
        lines = yaml_content.split('\n')
        cleaned_lines = []
        
        for i, line in enumerate(lines):
            stripped = line.strip()
            
            # Skip empty lines
            if not stripped:
                cleaned_lines.append(line)
                continue
            
            # Check for lines that look like standalone values without proper YAML structure
            # This often happens with IP addresses or hostnames in configuration files
            if ':' not in stripped and not stripped.startswith('-') and not stripped.startswith('#'):
                # Check if this looks like a problematic line pattern
                if (re.match(r'^\d+\.\d+\.\d+\.\d+\s+', stripped) or 
                    re.match(r'^[a-zA-Z0-9\-\.]+\s+[a-zA-Z0-9\-\.]+\s*$', stripped)):
                    
                    # Preserve leading whitespace but comment out the problematic line
                    leading_spaces = len(line) - len(line.lstrip())
                    if leading_spaces > 0:
                        # If it has indentation, it might be part of a multiline string - keep it
                        cleaned_lines.append(line)
                    else:
                        # Top-level problematic line - comment it out
                        cleaned_lines.append(f"# {line}")
                    continue
            
            # Keep all other lines as-is
            cleaned_lines.append(line)
        
        return '\n'.join(cleaned_lines)
    
    def _preprocess_yaml_content(self, yaml_content: str) -> str:
        """Preprocess YAML content to fix common parsing issues."""
        lines = yaml_content.split('\n')
        processed_lines = []
        
        in_multiline_string = False
        current_indent = 0
        
        for i, line in enumerate(lines):
            stripped = line.strip()
            
            # Skip empty lines
            if not stripped:
                processed_lines.append(line)
                continue
            
            # Check if this looks like a problematic line (IP address or hostname without proper YAML structure)
            if (re.match(r'^\d+\.\d+\.\d+\.\d+\s+', stripped) or 
                re.match(r'^[a-zA-Z0-9\-\.]+\s+[a-zA-Z0-9\-\.]+', stripped)) and ':' not in stripped:
                
                # This looks like a hosts file entry or similar - quote it as a string
                # Check if we're inside a multiline string context
                leading_spaces = len(line) - len(line.lstrip())
                
                # If this line doesn't have proper YAML structure, treat it as content
                if leading_spaces > 0:
                    # Preserve the line as-is if it's properly indented (likely part of a multiline string)
                    processed_lines.append(line)
                else:
                    # This is a top-level line that doesn't look like YAML - skip or comment it out
                    processed_lines.append(f"# {line}")
                continue
            
            # For all other lines, keep them as-is
            processed_lines.append(line)
        
        return '\n'.join(processed_lines)
    
    def _extract_value_from_path(self, doc: Dict, path: str) -> Any:
        """Extract value from nested dictionary using dot notation path."""
        try:
            # Handle array access like "cidrBlocks[0]"
            parts = self._parse_path(path)
            current = doc
            
            for part in parts:
                if isinstance(part, tuple):  # Array access
                    key, index = part
                    if key in current and isinstance(current[key], list):
                        if index < len(current[key]):
                            current = current[key][index]
                        else:
                            return None
                    else:
                        return None
                else:  # Regular key access
                    if isinstance(current, dict) and part in current:
                        current = current[part]
                    else:
                        return None
            
            return current
        except Exception:
            return None
    
    def _replace_value_in_doc(self, doc: Dict, path: str, new_value: str) -> None:
        """Replace value in nested dictionary using dot notation path."""
        try:
            parts = self._parse_path(path)
            current = doc
            
            # Navigate to the parent of the target
            for part in parts[:-1]:
                if isinstance(part, tuple):  # Array access
                    key, index = part
                    current = current[key][index]
                else:  # Regular key access
                    current = current[part]
            
            # Set the final value
            final_part = parts[-1]
            if isinstance(final_part, tuple):  # Array access
                key, index = final_part
                current[key][index] = new_value
            else:  # Regular key access
                current[final_part] = new_value
        except Exception:
            pass  # Ignore errors in replacement
    
    def _parse_path(self, path: str) -> List:
        """Parse a path like 'spec.cidrBlocks[0].name' into navigable parts."""
        parts = []
        segments = path.split('.')
        
        for segment in segments:
            # Check if this segment has array access
            if '[' in segment and ']' in segment:
                key = segment[:segment.index('[')]
                index_str = segment[segment.index('[') + 1:segment.index(']')]
                try:
                    index = int(index_str)
                    parts.append((key, index))
                except ValueError:
                    parts.append(segment)  # Invalid index, treat as regular key
            else:
                parts.append(segment)
        
        return parts

    def _parse_document_path(self, path: str) -> Tuple[Optional[str], str]:
        """Parse a path like 'Cluster.spec.infrastructureRef.name' into document type and actual path."""
        parts = path.split('.', 1)
        if len(parts) == 2:
            # Check if first part looks like a document type (capitalized)
            potential_doc_type = parts[0]
            if potential_doc_type and potential_doc_type[0].isupper():
                return potential_doc_type, parts[1]
        
        # No document type specified, return the whole path
        return None, path


class SpectroTerraformFormatter:
    """Spectro Terraform Formatter - Built-in YAML processing for Spectro Cloud"""
    
    def __init__(self, output_dir: str = None, backup: bool = None, format_tf: bool = True, only_yaml_format: bool = True):
        self.output_dir = output_dir or BUILTIN_TEMPLATING_CONFIG["output_dir"]
        self.backup = backup if backup is not None else BUILTIN_TEMPLATING_CONFIG["backup"]
        self.format_tf = format_tf
        self.only_yaml_format = only_yaml_format
    
    def process_file(self, tf_file_path: str) -> None:
        """Process a Terraform file with built-in Spectro Cloud configuration"""
        tf_file_path = Path(tf_file_path)
        
        if not tf_file_path.exists():
            raise FileNotFoundError(f"Terraform file not found: {tf_file_path}")
        
        print(f"Processing Terraform file: {tf_file_path}")
        
        # Read the Terraform file
        with open(tf_file_path, 'r', encoding='utf-8') as f:
            tf_content = f.read()
        
        # Create backup if requested
        if self.backup:
            backup_path = tf_file_path.with_suffix(tf_file_path.suffix + '.backup')
            with open(backup_path, 'w', encoding='utf-8') as f:
                f.write(tf_content)
            print(f"Created backup: {backup_path}")
        
        # Apply built-in YAML processing
        modified_content, created_files = self._apply_yaml_extraction(tf_content, tf_file_path)
        
        # Write the modified content back
        if modified_content != tf_content:
            with open(tf_file_path, 'w', encoding='utf-8') as f:
                f.write(modified_content)
            print(f"✓ Updated Terraform file: {tf_file_path}")
            
            # Format the Terraform file if requested
            if self.format_tf:
                self._format_terraform_file(tf_file_path)
        
        if created_files:
            print(f"\n✅ Successfully processed {len(created_files)} configurations")
            print("\nNext steps:")
            print("1. Review the generated configuration files")
            print("2. Modify as needed") 
            print("3. Run 'terraform plan' to verify changes")
            print("4. Commit files to version control")
        else:
            print("No configurations were processed")
    
    def _cleanup_output_dir(self, output_dir: Path) -> None:
        """Clean up and recreate the output directory for a fresh start"""
        if output_dir.exists():
            print(f"🧹 Cleaning up existing directory: {output_dir}")
            shutil.rmtree(output_dir)
        
        print(f"📁 Creating output directory: {output_dir}")
        output_dir.mkdir(parents=True, exist_ok=True)
    
    def _apply_yaml_extraction(self, tf_content: str, tf_file_path: Path) -> Tuple[str, List[str]]:
        """Extract YAML content from Terraform file using built-in configuration"""
        output_dir = Path(tf_file_path.parent / self.output_dir)
        self._cleanup_output_dir(output_dir)
        
        modified_content = tf_content
        created_files = []
        all_overrides = {}
        
        print(f"Applying built-in YAML extraction...")
        
        # Process each configured field (skip script configuration)
        script_config_keys = {"rules", "output_dir", "backup"}
        
        # Check if we should skip templating (only extract YAML)
        only_yaml_format = self.only_yaml_format
        
        for field_path, field_config in BUILTIN_TEMPLATING_CONFIG.items():
            # Skip script configuration fields
            if field_path in script_config_keys:
                continue
                
            print(f"Processing field: {field_path}")
            
            if field_path == "cloud_config.values":
                # Extract cloud_config values
                extractions = self._extract_cloud_config_values(tf_content)
                for resource_name, original_values, yaml_content in extractions:
                    yaml_file = output_dir / field_config["filename"].format(resource_name=resource_name)
                    
                    # Process templating if configured and not in only_yaml_format mode
                    final_yaml_content = yaml_content
                    resource_overrides = {}
                    
                    if not only_yaml_format and field_config.get("templating"):
                        templater = YAMLTemplater()
                        final_yaml_content, resource_overrides = templater.process_yaml(yaml_content, field_config["templating"])
                        print(f"    → Extracted {len(resource_overrides)} variables: {list(resource_overrides.keys())}")
                    elif only_yaml_format:
                        # In only_yaml_format mode, preserve the content exactly as-is
                        final_yaml_content = yaml_content
                        print(f"    → Skipping templating and preserving original format (only_yaml_format=True)")
                    
                    self._write_yaml_file(yaml_file, final_yaml_content)
                    created_files.append(str(yaml_file))
                    
                    # Create file reference (ensure forward slashes for Terraform compatibility)
                    relative_path = os.path.relpath(yaml_file, tf_file_path.parent)
                    if IS_WINDOWS:
                        relative_path = relative_path.replace('\\', '/')
                    file_reference = f'file("{relative_path}")'
                    
                    # Replace the values assignment
                    modified_content = self._replace_attribute_value(
                        modified_content, "values", original_values, file_reference
                    )
                    
                    # Store overrides for injection (only if not in only_yaml_format mode)
                    if not only_yaml_format and resource_overrides:
                        all_overrides[f"cloud_config_{resource_name}"] = resource_overrides
                    
                    print(f"    → Replaced cloud_config.values with {file_reference}")
                    
            elif field_path == "machine_pool.node_pool_config":
                # Extract machine_pool configs
                extractions = self._extract_machine_pool_configs(tf_content)
                for resource_name, pool_name, original_config, yaml_content in extractions:
                    yaml_file = output_dir / field_config["filename"].format(resource_name=resource_name, pool_name=pool_name)
                    
                    # Process templating if configured and not in only_yaml_format mode
                    final_yaml_content = yaml_content
                    resource_overrides = {}
                    
                    if not only_yaml_format and field_config.get("templating"):
                        templater = YAMLTemplater()
                        final_yaml_content, resource_overrides = templater.process_yaml(yaml_content, field_config["templating"])
                        print(f"    → Extracted {len(resource_overrides)} variables: {list(resource_overrides.keys())}")
                    elif only_yaml_format:
                        # In only_yaml_format mode, preserve the content exactly as-is
                        final_yaml_content = yaml_content
                        print(f"    → Skipping templating and preserving original format (only_yaml_format=True)")
                    
                    self._write_yaml_file(yaml_file, final_yaml_content)
                    created_files.append(str(yaml_file))
                    
                    # Create file reference (ensure forward slashes for Terraform compatibility)
                    relative_path = os.path.relpath(yaml_file, tf_file_path.parent)
                    if IS_WINDOWS:
                        relative_path = relative_path.replace('\\', '/')
                    file_reference = f'file("{relative_path}")'
                    
                    # Replace the node_pool_config assignment
                    modified_content = self._replace_attribute_value(
                        modified_content, "node_pool_config", original_config, file_reference
                    )
                    
                    # Store overrides for injection (only if not in only_yaml_format mode)
                    if not only_yaml_format and resource_overrides:
                        all_overrides[f"machine_pool_{resource_name}_{pool_name}"] = resource_overrides
                    
                    print(f"    → Replaced machine_pool.node_pool_config with {file_reference}")
        
        # Inject overrides into terraform file (only if not in only_yaml_format mode)
        if not only_yaml_format and all_overrides:
            modified_content = self._inject_overrides(modified_content, all_overrides)
            print(f"    → Injected {len(all_overrides)} override blocks")
        elif only_yaml_format:
            print(f"    → Skipping overrides injection (only_yaml_format=True)")
        
        # Post-process YAML files to fix preKubeadmCommands formatting
        if created_files:
            self._fix_prekubeadm_commands(created_files)
        
        return modified_content, created_files
    
    def _fix_prekubeadm_commands(self, created_files: List[str]) -> None:
        """Fix preKubeadmCommands formatting to use single-line echo commands with \n sequences"""
        print(f"\n🔧 Fixing preKubeadmCommands formatting in {len(created_files)} files...")
        
        fixed_count = 0
        for file_path in created_files:
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                original_content = content
                
                # Process the content line by line
                lines = content.split('\n')
                fixed_lines = []
                i = 0
                in_prekubeadm_section = False
                
                while i < len(lines):
                    line = lines[i]
                    
                    # Check if we're entering or leaving preKubeadmCommands section
                    if 'preKubeadmCommands:' in line:
                        in_prekubeadm_section = True
                        fixed_lines.append(line)
                        i += 1
                        continue
                    elif in_prekubeadm_section and line.strip() and not line.strip().startswith('-') and not line.startswith('    '):
                        # We've left the preKubeadmCommands section
                        in_prekubeadm_section = False
                    
                    # If we're in preKubeadmCommands section and find a multi-line echo command
                    if (in_prekubeadm_section and 
                        line.strip().startswith('- sudo echo -e "') and 
                        line.rstrip().endswith('\\')):
                        
                        # This is a multi-line echo command that needs fixing
                        echo_parts = []
                        
                        # Extract the first part (remove the leading format and trailing backslash)
                        first_part = line.strip()[len('- sudo echo -e "'):-1]  # Remove '- sudo echo -e "' and '\'
                        echo_parts.append(first_part)
                        i += 1
                        
                        # Collect all continuation lines
                        while i < len(lines) and lines[i].rstrip().endswith('\\'):
                            continuation = lines[i].strip()[:-1]  # Remove trailing backslash
                            echo_parts.append(continuation)
                            i += 1
                        
                        # Get the final line (without backslash but with closing quote and redirect)
                        if i < len(lines):
                            final_line = lines[i].strip()
                            if final_line.endswith('" >> /etc/hosts'):
                                final_part = final_line[:-len('" >> /etc/hosts')]
                                echo_parts.append(final_part)
                                i += 1
                            else:
                                # Fallback: take the whole final line and adjust
                                echo_parts.append(final_line.rstrip('"'))
                                i += 1
                        
                        # Reconstruct as single line with \n sequences
                        full_echo_content = '\\n'.join(echo_parts)
                        
                        # Get the original indentation
                        original_indent = len(line) - len(line.lstrip())
                        indent = ' ' * original_indent
                        
                        # Create the fixed single-line command
                        fixed_line = f'{indent}- sudo echo -e "{full_echo_content}" >> /etc/hosts'
                        fixed_lines.append(fixed_line)
                        
                    else:
                        # Regular line, keep as-is
                        fixed_lines.append(line)
                        i += 1
                
                # Join the fixed lines
                fixed_content = '\n'.join(fixed_lines)
                
                # Only write if content changed
                if fixed_content != original_content:
                    with open(file_path, 'w', encoding='utf-8') as f:
                        f.write(fixed_content)
                    fixed_count += 1
                    print(f"    ✅ Fixed preKubeadmCommands in {Path(file_path).name}")
                
            except Exception as e:
                print(f"    ❌ Error processing {Path(file_path).name}: {e}")
        
        if fixed_count > 0:
            print(f"    → Fixed preKubeadmCommands formatting in {fixed_count} files")
        else:
            print(f"    → No preKubeadmCommands fixes needed")
    
    def _extract_block_content(self, content: str, start_pos: int) -> str:
        """Extract block content from start_pos to matching closing brace"""
        brace_count = 1  # We start after the opening brace
        i = start_pos
        
        while i < len(content) and brace_count > 0:
            char = content[i]
            if char == '{':
                brace_count += 1
            elif char == '}':
                brace_count -= 1
            i += 1
        
        if brace_count == 0:
            return content[start_pos:i-1]  # Exclude the closing brace
        else:
            return content[start_pos:]  # Return everything if no matching brace found

    def _extract_cloud_config_values(self, tf_content: str) -> List[Tuple[str, str, str]]:
        """Extract cloud_config values from Terraform content"""
        extractions = []
        
        # Find all cloud_config blocks - need to handle nested braces properly
        cloud_config_pattern = r'cloud_config\s*\{'
        
        for cloud_match in re.finditer(cloud_config_pattern, tf_content, re.DOTALL):
            # Extract the block content manually by finding matching closing brace
            start_pos = cloud_match.end()  # Position after opening {
            cloud_block = self._extract_block_content(tf_content, start_pos)
            
            # Look for values attribute within this cloud_config block
            values_pattern = r'values\s*=\s*'
            values_match = re.search(values_pattern, cloud_block, re.DOTALL)
            
            if values_match:
                # Extract the quoted string starting from after the = sign
                quote_start = values_match.end()
                quoted_string = self._extract_complete_quoted_string(cloud_block, quote_start)
                
                # Skip if already using file() function
                if 'file(' in quoted_string:
                    print(f"    Skipping already processed cloud_config.values: {quoted_string[:50]}...")
                    continue
                
                # Find the resource name by looking backwards
                before_cloud_config = tf_content[:cloud_match.start()]
                resource_pattern = r'resource\s+"[^"]+"\s+"([^"]+)"\s*\{'
                resource_matches = list(re.finditer(resource_pattern, before_cloud_config, re.DOTALL))
                
                if resource_matches:
                    resource_name = resource_matches[-1].group(1)
                    # Unescape the string to get actual YAML content
                    yaml_content = self._unescape_terraform_string(quoted_string)
                    print(f"  Found cloud_config.values for resource: {resource_name} (length: {len(quoted_string)} chars)")
                    extractions.append((resource_name, quoted_string, yaml_content))
                else:
                    print(f"    Could not find resource name for cloud_config.values")
        
        return extractions
    
    def _extract_machine_pool_configs(self, tf_content: str) -> List[Tuple[str, str, str, str]]:
        """Extract machine_pool configurations from Terraform content"""
        extractions = []
        
        # Find all machine_pool blocks - need to handle nested braces properly
        machine_pool_pattern = r'machine_pool\s*\{'
        
        for pool_match in re.finditer(machine_pool_pattern, tf_content, re.DOTALL):
            # Extract the block content manually by finding matching closing brace
            start_pos = pool_match.end()  # Position after opening {
            pool_block = self._extract_block_content(tf_content, start_pos)
            
            # Look for node_pool_config attribute within this machine_pool block
            config_pattern = r'node_pool_config\s*=\s*'
            config_match = re.search(config_pattern, pool_block, re.DOTALL)
            
            if config_match:
                # Extract the quoted string starting from after the = sign
                quote_start = config_match.end()
                quoted_string = self._extract_complete_quoted_string(pool_block, quote_start)
                
                # Skip if already using file() function
                if 'file(' in quoted_string:
                    print(f"    Skipping already processed machine_pool config: {quoted_string[:50]}...")
                    continue
                
                # Find the resource name by looking backwards
                before_machine_pool = tf_content[:pool_match.start()]
                resource_pattern = r'resource\s+"[^"]+"\s+"([^"]+)"\s*\{'
                resource_matches = list(re.finditer(resource_pattern, before_machine_pool, re.DOTALL))
                
                if resource_matches:
                    resource_name = resource_matches[-1].group(1)
                    
                    # Start with a basic fallback name (will be overridden by YAML extraction if available)
                    if 'control_plane = true' in pool_block:
                        pool_name = f"{resource_name}-cp"  # Fallback for control plane
                    else:
                        pool_name = f"{resource_name}-worker"  # Fallback for worker
                    
                    # Unescape the string to get actual YAML content
                    yaml_content = self._unescape_terraform_string(quoted_string)
                    
                    # Try to extract actual name from YAML content for better naming (but ensure uniqueness)
                    if 'name:' in yaml_content:
                        name_match = re.search(r'name:\s*([^\n\r]+)', yaml_content)
                        if name_match:
                            extracted_name = name_match.group(1).strip()
                            # Clean up template variables and use as pool name if valid
                            cleaned_name = re.sub(r'\s*\{\{.*?\}\}', '', extracted_name).strip()
                            if cleaned_name and not cleaned_name.startswith('${'):
                                # Check if this name is already used and add counter if needed
                                already_used_names = [name for _, name, _, _ in extractions]
                                if cleaned_name not in already_used_names:
                                    pool_name = cleaned_name  # Use extracted name if unique
                                else:
                                    # Name already exists, add counter
                                    counter = 2
                                    while f"{cleaned_name}_{counter}" in already_used_names:
                                        counter += 1
                                    pool_name = f"{cleaned_name}_{counter}"
                    
                    print(f"  Found machine_pool.node_pool_config for resource: {resource_name}, pool: {pool_name} (length: {len(quoted_string)} chars)")
                    extractions.append((resource_name, pool_name, quoted_string, yaml_content))
                else:
                    print(f"    Could not find resource name for machine_pool.node_pool_config")
        
        return extractions
    
    def _extract_complete_quoted_string(self, content: str, start_pos: int) -> str:
        """Extract a complete quoted string starting from start_pos"""
        # Skip whitespace to find the opening quote
        pos = start_pos
        while pos < len(content) and content[pos].isspace():
            pos += 1
        
        if pos >= len(content) or content[pos] != '"':
            return ""
        
        # Start after the opening quote
        result = []
        pos += 1
        
        while pos < len(content):
            char = content[pos]
            
            if char == '"':
                # Check if it's escaped
                if pos > 0 and content[pos - 1] == '\\':
                    # Count consecutive backslashes
                    num_backslashes = 0
                    check_pos = pos - 1
                    while check_pos >= 0 and content[check_pos] == '\\':
                        num_backslashes += 1
                        check_pos -= 1
                    
                    # If odd number of backslashes, quote is escaped
                    if num_backslashes % 2 == 1:
                        result.append(char)
                        pos += 1
                        continue
                
                # Found closing quote
                return '"' + ''.join(result) + '"'
            else:
                result.append(char)
            
            pos += 1
        
        # If we reach here, no closing quote found
        return ""
    
    def _unescape_terraform_string(self, quoted_string: str) -> str:
        """Convert Terraform escaped string to actual content"""
        # Remove outer quotes
        if quoted_string.startswith('"') and quoted_string.endswith('"'):
            content = quoted_string[1:-1]
        else:
            content = quoted_string
        
        # For normal content, do standard unescaping
        content = content.replace('\\"', '"')
        content = content.replace('\\n', '\n')
        content = content.replace('\\r', '\r')
        content = content.replace('\\t', '\t')
        content = content.replace('\\\\', '\\')
        
        return content
    
    def _write_yaml_file(self, yaml_file: Path, content: str) -> None:
        """Write YAML content to file"""
        with open(yaml_file, 'w', encoding='utf-8') as f:
            f.write(content)
    
    def _replace_attribute_value(self, content: str, attribute_name: str, original_value: str, new_value: str) -> str:
        """Replace an attribute value in terraform content"""
        # More robust regex-based replacement that finds the exact attribute assignment
        # This handles cases where the same quoted string might appear elsewhere
        
        # First, try to find the attribute assignment specifically
        # Pattern: attribute_name = "quoted_string"
        escaped_original = re.escape(original_value)
        pattern = f'({attribute_name}\\s*=\\s*){escaped_original}'
        
        replacement = f'\\1{new_value}'
        modified_content = re.sub(pattern, replacement, content, flags=re.DOTALL)
        
        # Check if replacement was successful
        if modified_content != content:
            print(f"      ✓ Successfully replaced {attribute_name} using regex pattern")
            return modified_content
        
        # Fallback: try direct string replacement if regex didn't work
        if original_value in content:
            print(f"      ✓ Successfully replaced {attribute_name} using direct string replacement")
            return content.replace(original_value, new_value)
        
        # If both methods failed, log the issue but continue
        print(f"      ⚠️  Could not replace {attribute_name} = {original_value[:50]}...")
        return content
    
    def _inject_overrides(self, tf_content: str, all_overrides: Dict[str, Dict[str, Any]]) -> str:
        """Inject overrides blocks into terraform configuration"""
        if not all_overrides:
            return tf_content
            
        modified_content = tf_content
        
        # Process each resource's overrides separately 
        for key, overrides in all_overrides.items():
            if key.startswith('cloud_config_'):
                resource_name = key.replace('cloud_config_', '')
                modified_content = self._inject_cloud_config_overrides_for_resource(modified_content, overrides, resource_name)
            elif key.startswith('machine_pool_'):
                # Extract resource name and pool name from key like "machine_pool_capi_cluster_worker-pool"
                parts = key.replace('machine_pool_', '').split('_', 1)
                if len(parts) >= 2:
                    resource_name = parts[0]
                    pool_name = parts[1]
                    modified_content = self._inject_machine_pool_overrides_for_resource(modified_content, overrides, resource_name, pool_name)
        
        return modified_content
    
    def _inject_cloud_config_overrides_for_resource(self, tf_content: str, overrides: Dict[str, Any], resource_name: str) -> str:
        """Inject overrides into cloud_config block for a specific resource"""
        if not overrides:
            return tf_content
            
        # Format overrides
        overrides_lines = []
        for key, value in overrides.items():
            if isinstance(value, str):
                overrides_lines.append(f'      {key} = "{value}"')
            else:
                overrides_lines.append(f'      {key} = {value}')
        
        new_overrides_block = "overrides = {\n" + "\n".join(overrides_lines) + "\n    }"
        
        # Use a more targeted approach - find cloud_config blocks that contain file() references
        # and replace their overrides
        def replace_cloud_config_overrides(match):
            full_block = match.group(0)
            cloud_block = match.group(1)
            
            # Only process blocks that have file() reference (were processed)
            if 'file(' not in cloud_block:
                return full_block
            
            # Replace existing overrides or add new ones
            if 'overrides = {}' in cloud_block:
                # Replace empty overrides
                updated_cloud_block = cloud_block.replace('overrides = {}', new_overrides_block)
            elif re.search(r'overrides\s*=\s*\{[^}]*\}', cloud_block):
                # Replace existing non-empty overrides
                updated_cloud_block = re.sub(r'overrides\s*=\s*\{[^}]*\}', new_overrides_block, cloud_block, flags=re.DOTALL)
            else:
                # Add overrides at the beginning of the cloud_config block
                updated_cloud_block = f"\n    {new_overrides_block}\n{cloud_block}"
            
            return f"cloud_config {{{updated_cloud_block}}}"
        
        # Apply the replacement to cloud_config blocks that contain file() references
        # Use the same block extraction method we use elsewhere for robustness
        cloud_config_pattern = r'cloud_config\s*\{'
        
        modified_content = tf_content
        for cloud_match in re.finditer(cloud_config_pattern, tf_content, re.DOTALL):
            start_pos = cloud_match.end()
            cloud_block = self._extract_block_content(tf_content, start_pos)
            
            # Only process blocks that have file() reference (were processed)
            if 'file(' in cloud_block:
                # Replace existing overrides or add new ones
                if 'overrides = {}' in cloud_block:
                    updated_cloud_block = cloud_block.replace('overrides = {}', new_overrides_block)
                elif re.search(r'overrides\s*=\s*\{[^}]*\}', cloud_block):
                    updated_cloud_block = re.sub(r'overrides\s*=\s*\{[^}]*\}', new_overrides_block, cloud_block, flags=re.DOTALL)
                else:
                    updated_cloud_block = f"\n    {new_overrides_block}\n{cloud_block}"
                
                # Replace in the content
                full_original = f"cloud_config {{{cloud_block}}}"
                full_updated = f"cloud_config {{{updated_cloud_block}}}"
                modified_content = modified_content.replace(full_original, full_updated)
        
        return modified_content
    
    def _inject_machine_pool_overrides_for_resource(self, tf_content: str, overrides: Dict[str, Any], resource_name: str, pool_name: str) -> str:
        """Inject overrides into machine_pool block for a specific resource and pool"""
        if not overrides:
            return tf_content
            
        # Format overrides
        overrides_lines = []
        for key, value in overrides.items():
            if isinstance(value, str):
                overrides_lines.append(f'      {key} = "{value}"')
            else:
                overrides_lines.append(f'      {key} = {value}')
        
        new_overrides_block = "overrides = {\n" + "\n".join(overrides_lines) + "\n    }"
        
        # Expected file reference for this specific pool
        expected_file_ref = f'{resource_name}_{pool_name}_config.yaml'
        
        # Apply the replacement to machine_pool blocks that contain the specific file reference
        machine_pool_pattern = r'machine_pool\s*\{'
        
        modified_content = tf_content
        for pool_match in re.finditer(machine_pool_pattern, tf_content, re.DOTALL):
            start_pos = pool_match.end()
            pool_block = self._extract_block_content(tf_content, start_pos)
            
            # Only process the specific block that has this pool's file reference
            if expected_file_ref in pool_block:
                # Replace existing overrides or add new ones
                if 'overrides = {}' in pool_block:
                    updated_pool_block = pool_block.replace('overrides = {}', new_overrides_block)
                elif re.search(r'overrides\s*=\s*\{[^}]*\}', pool_block):
                    updated_pool_block = re.sub(r'overrides\s*=\s*\{[^}]*\}', new_overrides_block, pool_block, flags=re.DOTALL)
                else:
                    updated_pool_block = f"\n    {new_overrides_block}\n{pool_block}"
                
                # Replace in the content
                full_original = f"machine_pool {{{pool_block}}}"
                full_updated = f"machine_pool {{{updated_pool_block}}}"
                modified_content = modified_content.replace(full_original, full_updated)
                break  # Only process the matching pool
        
        return modified_content
    
    def _get_terraform_commands(self) -> List[str]:
        """Get list of terraform commands to try for the current platform"""
        if IS_WINDOWS:
            return ['terraform.exe', 'terraform']
        else:
            return ['terraform']
    
    def _format_terraform_file(self, tf_path: Path) -> None:
        """Format the Terraform file using terraform fmt (cross-platform)"""
        print(f"📝 Formatting Terraform file...")
        
        commands_to_try = self._get_terraform_commands()
        
        for terraform_cmd in commands_to_try:
            try:
                result = subprocess.run(
                    [terraform_cmd, 'fmt', str(tf_path)],
                    cwd=tf_path.parent,
                    capture_output=True,
                    text=True,
                    timeout=30,
                    shell=IS_WINDOWS  # Use shell on Windows for better PATH resolution
                )
                
                if result.returncode == 0:
                    print("✅ Terraform file formatted successfully")
                    return
                else:
                    print(f"⚠️  Terraform fmt warning: {result.stderr.strip()}")
                    return
                    
            except subprocess.TimeoutExpired:
                print(f"⚠️  Terraform fmt timed out with command: {terraform_cmd}")
                return
            except FileNotFoundError:
                # Try next command in the list
                continue
            except Exception as e:
                print(f"⚠️  Error running {terraform_cmd}: {e}")
                continue
        
        # If we get here, none of the commands worked
        if IS_WINDOWS:
            print("⚠️  terraform command not found - ensure terraform.exe is in your PATH")
            print("    Download from: https://www.terraform.io/downloads.html")
        else:
            print("⚠️  terraform command not found - skipping formatting")


def main():
    """Main entry point"""
    # Platform-specific examples
    if IS_WINDOWS:
        platform_examples = """
Windows Examples:
  # Using batch wrapper (recommended)
  spectro-tf-format.bat generated.tf
  
  # Using PowerShell wrapper
  .\\spectro-tf-format.ps1 generated.tf
  
  # Direct Python execution
  python spectro-tf-format generated.tf"""
    else:
        platform_examples = """
Unix/Linux Examples:
  # Make executable first
  chmod +x spectro-tf-format
  
  # Run directly
  ./spectro-tf-format generated.tf"""

    parser = argparse.ArgumentParser(
        description='Spectro Terraform Formatter - Built-in YAML processing for Spectro Cloud (Cross-Platform)',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=f"""
{platform_examples}

General Examples:
  # Process terraform file with defaults (YAML extraction only)
  spectro-tf-format generated.tf
  
  # Enable templating and overrides processing
  spectro-tf-format --with-templating generated.tf
  
  # Explicit YAML-only mode (same as default)
  spectro-tf-format --only-yaml-format generated.tf
  
  # Use custom output directory (default: {BUILTIN_TEMPLATING_CONFIG["output_dir"]})
  spectro-tf-format --output-dir=cluster_configs_yaml generated.tf
  
  # Skip backup creation (default: backup={BUILTIN_TEMPLATING_CONFIG["backup"]})
  spectro-tf-format --no-backup generated.tf
  
  # Skip terraform formatting
  spectro-tf-format --no-format generated.tf

Built-in Defaults:
  Mode: YAML extraction only (use --with-templating for full processing)
  Output Directory: {BUILTIN_TEMPLATING_CONFIG["output_dir"]}
  Backup Files: {BUILTIN_TEMPLATING_CONFIG["backup"]}
  Rules: {', '.join(BUILTIN_TEMPLATING_CONFIG["rules"])}
  Platform: {platform.system()} ({platform.machine()})
        """
    )
    
    parser.add_argument('terraform_file', help='Terraform file to process')
    parser.add_argument('--output-dir', '-o', default=BUILTIN_TEMPLATING_CONFIG["output_dir"], help=f'Output directory for generated files (default: {BUILTIN_TEMPLATING_CONFIG["output_dir"]})')
    parser.add_argument('--no-backup', action='store_true', help='Skip creating backup file')
    parser.add_argument('--no-format', action='store_true', help='Skip automatic terraform fmt formatting')
    parser.add_argument('--only-yaml-format', action='store_true', help='Only extract YAML to files without templating or overrides (DEFAULT)')
    parser.add_argument('--with-templating', action='store_true', help='Enable templating and overrides processing')
    
    args = parser.parse_args()
    
    if not args.terraform_file:
        print("Error: Terraform file is required", file=sys.stderr)
        parser.print_help()
        sys.exit(1)
    
    try:
        # Use built-in defaults if not overridden
        backup = not args.no_backup if args.no_backup else BUILTIN_TEMPLATING_CONFIG["backup"]
        
        # Determine only_yaml_format based on flags
        if args.with_templating and args.only_yaml_format:
            print("Error: Cannot specify both --with-templating and --only-yaml-format", file=sys.stderr)
            sys.exit(1)
        elif args.with_templating:
            only_yaml_format = False  # Enable templating
        else:
            only_yaml_format = True   # Default: YAML-only mode
        
        formatter = SpectroTerraformFormatter(
            output_dir=args.output_dir,
            backup=backup,
            format_tf=not args.no_format,
            only_yaml_format=only_yaml_format
        )
        formatter.process_file(args.terraform_file)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main() 