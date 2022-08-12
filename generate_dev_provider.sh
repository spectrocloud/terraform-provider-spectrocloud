#set -x

PROVIDER_VERSION=$1

if [ -z "$PROVIDER_VERSION" ]
then
  echo "No provider version is provided. Thus generating provider with version 100.100.100."
  echo "Initialize provider in hcl file as >= on your current version. Eg. >= 0.7.7"
  PROVIDER_VERSION=100.100.100
fi

echo "Generating spectrocloud local provider with version $PROVIDER_VERSION"

OS=$(echo $(uname -s) | tr '[:upper:]' '[:lower:]')
if [ $(uname -m) == "x86_64" ]; then
  OS_ARCH="amd64"
elif [ $(uname -m) == "i386" ]; then
  OS_ARCH="386"
elif [ $(uname -m) == "i686" ]; then
  OS_ARCH="386"
elif [ $(uname -m) == "arm64" ]; then
  OS_ARCH="arm64"
fi
OS_VERSION=${OS}_${OS_ARCH}

function downloadProviderFromRegistry() {
  PROVIDER_NAME=$1
  PROVIDER_V=$2
  PROVIDER_MAINTAINER=$3
  if [ -z "$PROVIDER_MAINTAINER" ]
  then
    PROVIDER_MAINTAINER="hashicorp"
  fi

  echo "Downloading: $PROVIDER_NAME, $PROVIDER_V, $OS_VERSION"

  TERRAFORM_DIR=.terraform.d/plugins/registry.terraform.io/"${PROVIDER_MAINTAINER}"
  HASHICORP_RELEASE_DOMAIN=https://releases.hashicorp.com

  PROVIDER_RUL=$HASHICORP_RELEASE_DOMAIN/terraform-provider-"$PROVIDER_NAME"/"$PROVIDER_V"/terraform-provider-"$PROVIDER_NAME"_"$PROVIDER_V"_"$OS_VERSION".zip
  wget "$PROVIDER_RUL" -O provider.zip &&
    unzip provider.zip &&
    chmod +x terraform-provider-"$PROVIDER_NAME"_* &&
    mkdir -p $TERRAFORM_DIR/"$PROVIDER_NAME"/"$PROVIDER_V"/"$OS_VERSION" &&
    mv terraform-provider-"$PROVIDER_NAME"_* $TERRAFORM_DIR/"$PROVIDER_NAME"/"$PROVIDER_V"/"$OS_VERSION" &&
  rm -f provider.zip
}

function generateSpectrocloudProvider() {
  rm -rf dist
  goreleaser build --skip-validate --rm-dist -f dev/dev-goreleaser.yml
}

function generateProvidersPlugins() {
  # copy spectrocloud generated provider
  chmod +x dist/terraform-provider-spectrocloud_"$OS_VERSION"/terraform-provider-spectrocloud_"$OS_VERSION"
  mkdir -p providers/plugins/registry.terraform.io/spectrocloud/spectrocloud/"$PROVIDER_VERSION"/"$OS_VERSION"
  cp dist/terraform-provider-spectrocloud_"$OS_VERSION"/terraform-provider-spectrocloud_"$OS_VERSION" providers/plugins/registry.terraform.io/spectrocloud/spectrocloud/"$PROVIDER_VERSION"/"$OS_VERSION"/

  # copy other downloded providers
  cp -R .terraform.d/plugins providers/
  rm -rf .terraform.d
}

function prepare() {
  rm -rf dist providers
}

function echoInitCmd() {
  echo "$(tput setaf 4) ✅ Run below command to initialize terraform with local generated provider ($PROVIDER_VERSION):"
  echo "$(tput setaf 1) terraform init --plugin-dir $(pwd)/providers/plugins"
}

prepare
generateSpectrocloudProvider
downloadProviderFromRegistry random 3.1.0
downloadProviderFromRegistry template 2.2.0
downloadProviderFromRegistry null 3.1.0
downloadProviderFromRegistry local 2.1.0
generateProvidersPlugins
echoInitCmd
