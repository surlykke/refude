#!/bin/bash
cd `dirname $0`

BIN_DIR=${HOME}/.local/bin
BASH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/bash-completion/completions
ZSH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/zsh/completions
FISH_COMPLETION_DIR=${XDG_CONFIG_HOME:-${HOME}/.config}/fish/completions
mkdir -p $BIN_DIR $BASH_COMPLETION_DIR $ZSH_COMPLETION_DIR $FISH_COMPLETION_DIR
APPS="RefudeDesktopService RefudeIconService RefudePowerService RefudeWmService RefudeXdgOpen RefudeStatusNotifierService RefudeNotificationService RefudeProxy"


for dir in $APPS; do
	(cd $dir && echo "Building $dir" && go install) || exit 1
done

cp scripts/bin/* ${BIN_DIR}
cp scripts/completions/bash/* ${BASH_COMPLETION_DIR}
cp scripts/completions/zsh/* ${ZSH_COMPLETION_DIR}
cp scripts/completions/fish/* ${FISH_COMPLETION_DIR}
