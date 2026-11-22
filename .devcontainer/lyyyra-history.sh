export HISTFILE=/home/vscode/.bash_history_data/.bash_history
shopt -s histappend
export HISTSIZE=10000
export HISTFILESIZE=20000
export PROMPT_COMMAND="history -a; history -n; ${PROMPT_COMMAND:-}"
