# Common set of functions
# Error check is done with set -e command . Build will fail if any of the commands fail
set -x

DATE=$(date '+%Y%m%d')

print_step() {
        text_val=$1
        set +x
        echo " "
        echo "###################################################
#  ${text_val}
###################################################"
        echo " "
        set -x
}

set_image_tag() {
        IMG_TAG="latest"
        RELEASE_FLAG=no

        if [[ ${JOB_TYPE} == 'presubmit' ]]; then
            VERSION_SUFFIX="-dev"
            IMG_LOC='pr'
            IMG_TAG=${PULL_NUMBER}
            PROD_BUILD_ID=${IMG_TAG}
            IMG_PATH=spectro-images/${IMG_LOC}
            GS_ARTIFACT_LOC=gs://spectro-prow-artifacts/pr-logs/pull/${REPO_OWNER}_${REPO_NAME}/${PULL_NUMBER}/${JOB_NAME}/${BUILD_NUMBER}/artifacts
            GS_COMPLIANCE_LOC=gs://spectro-prow-artifacts/pr-logs/pull/${REPO_OWNER}_${REPO_NAME}/${PULL_NUMBER}/${JOB_NAME}/${BUILD_NUMBER}/artifacts/compliance
        fi
        if [[ ${SPECTRO_FIPS} ]] && [[ ${SPECTRO_FIPS}  == "yes" ]]; then
           IMG_LOC=${IMG_LOC}-fips
        fi
        if [[ ${JOB_TYPE} == 'periodic' ]]; then
            VERSION_SUFFIX="-$(date +%y%m%d)"
            IMG_LOC='daily'
            IMG_TAG=$(date +%Y%m%d.%H%M)
            PROD_BUILD_ID=${IMG_TAG}
            IMG_PATH=spectro-images/${IMG_LOC}
            FLAGS=enable-onprem-public-clouds
            GS_ARTIFACT_LOC=gs://spectro-prow-artifacts/logs/${JOB_NAME}/${BUILD_NUMBER}/artifacts
            GS_COMPLIANCE_LOC=gs://spectro-prow-artifacts/compliance/$DATE
        fi
        if [[ ${SPECTRO_RELEASE} ]] && [[ ${SPECTRO_RELEASE} == "yes" ]]; then
            RELEASE_FLAG=yes
            export VERSION_SUFFIX=""
            IMG_LOC='release'
            IMG_TAG=$(make version)
            PROD_BUILD_ID=$(date +%Y%m%d.%H%M)
            IMG_PATH=spectro-images-client/${IMG_LOC}
            OVERLAY=overlays/release
            DOCKER_REGISTRY=${DOCKER_REGISTRY_CLIENT}
            # Local artifacts still get copied to pr location. Release artifacts get copied as part of copy_release_artifacts function
            GS_ARTIFACT_LOC=gs://spectro-prow-artifacts/pr-logs/pull/${REPO_OWNER}_${REPO_NAME}/${PULL_NUMBER}/${JOB_NAME}/${BUILD_NUMBER}/artifacts
            GS_COMPLIANCE_LOC=updated_in_set_release
        fi
        export PROD_BUILD_ID
        export IMG_PATH
        export IMG_TAG
        export VERSION_SUFFIX
        export PROD_VERSION=$(make version)
}

commenter() {
        export GITHUB_TOKEN=$ACCESS_TOKEN_PWD
        export GITHUB_OWNER=$REPO_OWNER
        export GITHUB_REPO=$REPO_NAME
        export GITHUB_COMMENT_TYPE=pr
        export GITHUB_PR_ISSUE_NUMBER=$PULL_NUMBER
        export GITHUB_COMMENT_FORMAT="Build logs for Job ${JOB_NAME} can be found here: {{.}}"
        export GITHUB_COMMENT="http://mayflower.spectrocloud.com/log?job=${JOB_NAME}&id=${BUILD_NUMBER}"
        github-commenter
}

run_unit_tests() {
   cd ../spectrocloud
   go test -v ./       

}


export REPO_NAME=terraform-provider-spectrocloud
set_image_tag
