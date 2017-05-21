#!/bin/bash
cd `dirname $0`

BIN_DIR=${HOME}/.local/bin
BASH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/bash-completion/completions
ZSH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/zsh/completions
mkdir -p $BIN_DIR $BASH_COMPLETION_DIR $ZSH_COMPLETION_DIR

for dir in RefudeDesktopService RefudeIconService RefudePowerService RefudeWmService RefudeXdgOpen; do
	(cd $dir && echo "Building $dir" && go install) || exit 1
done

cp scripts/runRefude.sh scripts/RefudeGET ${BIN_DIR}
cp scripts/RefudeGET.bash_completion ${BASH_COMPLETION_DIR}/RefudeGET
cp scripts/RefudeGET.zsh_completion ${ZSH_COMPLETION_DIR}/_RefudeGET
