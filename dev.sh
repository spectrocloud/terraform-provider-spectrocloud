export Q=$HOME/projects/spectro/terraform-provider-spectrocloud
alias f="(cd $Q && make)"

r() {
    local var="$1"
    # remove leading whitespace characters
    var="${var#"${var%%[![:space:]]*}"}"
    # remove trailing whitespace characters
    var="${var%"${var##*[![:space:]]}"}"
    #export TF_REATTACH_PROVIDERS="$var"
    export TF_REATTACH_PROVIDERS=$var
    echo "TF_REATTACH_PROVIDERS=${TF_REATTACH_PROVIDERS}"
}

#r() {
#  #spectrocloud.com/spectrocloud/spectrocloud":{"Protocol":"grpc","Pid":1490,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/bz/fhn9jmcj77j1rwmy65g1klsh0000gn/T/plugin390280241"}}}
#  export TF_REATTACH_PROVIDERS=
#}
