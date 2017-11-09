#!/bin/bash
# Author: Eli Qiao <qiaoliyong@gmail.com>

ret=$?
files=$(git diff HEAD --stat | awk '{if ($1 ~ /\.go$/) {print $1}}')
arr=($files)
# hacking command list
# don't check shift
cmds=("golint" "go fmt" "go tool vet -shift=false")

function hack {
    local cmd=$1
    local file=$2
    local rev

    rev=$($cmd "$file")

    if [[ $? -ne 0 ]] || [[ ! -z $rev ]]; then
        if [[ ! -z $rev ]]; then
            echo -e "\033[1;91m${rev} \033[0m"
        else
            # FIXME go tool vet does not give output
            echo -e "\033[1;91mgo tool vet $file failed \033[0m"
        fi
    else
        echo "0"
    fi
}

function do_check {
    local f=$1
    local rev
    if [ -f "${f}" ]; then
        for ((i = 0; i < ${#cmds[@]}; i++)) do

            rev=$(hack "${cmds[$i]}" $f)

            if [[ ! -z ${rev} && ${rev} != "0" ]]; then
                echo $rev
                RET=-1
            fi
        done
    fi
}

# global variable for checking result
RET=0

if [ $# -eq 1 ] && [ "$1" == "-f" ]; then
    echo "do full code checking ..."
    find ./ | grep -v vendor | grep -v .git | grep -v test |grep ".go"  > tmp
    while IFS='' read -r line || [[ -n "$line" ]]; do
        do_check $line
    done < "tmp"
    # echo "Total code lines:"
    # wc -l `find ./ | grep -v vendor | grep -v .git | grep -v test |grep ".go"`

    rm "tmp"
else
    for f in "${arr[@]}"
    do
        do_check $f
    done
fi

if [[ $RET -ne 0 ]]; then
    echo ":( <<< Please address coding style mistakes."
else
    echo ":) >>> No errors for coding style"
fi

exit $RET
