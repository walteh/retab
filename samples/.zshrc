##############################################
# Custom plugins
##############################################

custom_plugins=(
	# zsh-github-copilot:gopilot
	# github.com/hadenlabs/zsh-k9s
	# github.com/hadenlabs/zsh-core
	github.com/MichaelAquilina/zsh-you-should-use
	github.com/joshskidmore/zsh-fzf-history-search
)

mkdir -p "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins"

for plugin in "${custom_plugins[@]}"; do
	plugin_name=$(basename "$plugin")
	plugin_url=$plugin
	if [ ! -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/$plugin_name" ]; then
		git clone "https://$plugin_url" "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/$plugin_name"
	fi
done

##############################################
# Completions
##############################################

completions=(
	k9s:'k9s completion zsh'
	kn:'kn completion zsh'
	func:'func completion zsh'
	kn-func:'kn func completion zsh'
)

mkdir -p "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/completions"

for completion in "${completions[@]}"; do
	name=$(echo "$completion" | cut -d':' -f1)
	command=$(echo "$completion" | cut -d':' -f2)
	if [ ! -f "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/completions/_$name" ]; then
		if ! command -v "$name" &> /dev/null; then
			continue
		fi
		echo "Installing $name completions to ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/completions/_$name ($command)"
		bash -c "$command" > "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/completions/_$name"
	fi
done

##############################################
# zsh global plugins
##############################################

# If you come from bash you might have to change your $PATH.
# export PATH=$HOME/bin:$HOME/.local/bin:/usr/local/bin:$PATH

# source /opt/homebrew/share/zsh-autocomplete/zsh-autocomplete.plugin.zsh
if type brew &> /dev/null; then
	source $(brew --prefix)/share/zsh-autosuggestions/zsh-autosuggestions.zsh
	# source $(brew --prefix)/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
	source $(brew --prefix)/share/zsh-fast-syntax-highlighting/fast-syntax-highlighting.plugin.zsh

	FPATH=$(brew --prefix)/share/zsh-completions:$FPATH
	FPATH=$(brew --prefix)/share/zsh/site-functions:$FPATH
	autoload -Uz compinit
	compinit -i
fi

##### for github copilot suggections #####

# if [ ! -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-github-copilot" ]; then
#   git clone https://github.com/loiccoyle/zsh-github-copilot "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-github-copilot"
#   gh extension install github/gh-copilot --force
# fi
# bindkey '¿' zsh_gh_copilot_explain  # bind Option+shift+\ to explain
# bindkey '÷' zsh_gh_copilot_suggest  # bind Option+\ to suggest

##### for kubectl context #####

# autoload -U colors; colors
# source /opt/homebrew/etc/zsh-kubectl-prompt/kubectl.zsh
# RPROMPT='%{$fg[blue]%}($ZSH_KUBECTL_PROMPT)%{$reset_color%}'

# Path to your Oh My Zsh installation.
export ZSH="$HOME/.oh-my-zsh"

# Set name of the theme to load --- if set to "random", it will
# load a random theme each time Oh My Zsh is loaded, in which case,
# to know which specific one was loaded, run: echo $RANDOM_THEME
# See https://github.com/ohmyzsh/ohmyzsh/wiki/Themes
ZSH_THEME="robbyrussell"

# Set list of themes to pick from when loading at random
# Setting this variable when ZSH_THEME=random will cause zsh to load
# a theme from this variable instead of looking in $ZSH/themes/
# If set to an empty array, this variable will have no effect.
# ZSH_THEME_RANDOM_CANDIDATES=( "robbyrussell" "agnoster" )

# Uncomment the following line to use case-sensitive completion.
# CASE_SENSITIVE="true"

# Uncomment the following line to use hyphen-insensitive completion.
# Case-sensitive completion must be off. _ and - will be interchangeable.
HYPHEN_INSENSITIVE="true"

# Uncomment one of the following lines to change the auto-update behavior
# zstyle ':omz:update' mode disabled  # disable automatic updates
zstyle ':omz:update' mode auto # update automatically without asking
# zstyle ':omz:update' mode reminder  # just remind me to update when it's time

# Uncomment the following line to change how often to auto-update (in days).
# zstyle ':omz:update' frequency 13

# Uncomment the following line if pasting URLs and other text is messed up.
# DISABLE_MAGIC_FUNCTIONS="true"

# Uncomment the following line to disable colors in ls.
# DISABLE_LS_COLORS="true"

# Uncomment the following line to disable auto-setting terminal title.
# DISABLE_AUTO_TITLE="true"

# Uncomment the following line to enable command auto-correction.
# ENABLE_CORRECTION="true"

# Uncomment the following line to display red dots whilst waiting for completion.
# You can also set it to another string to have that shown instead of the default red dots.
# e.g. COMPLETION_WAITING_DOTS="%F{yellow}waiting...%f"
# Caution: this setting can cause issues with multiline prompts in zsh < 5.7.1 (see #5765)
COMPLETION_WAITING_DOTS="true"

# Uncomment the following line if you want to disable marking untracked files
# under VCS as dirty. This makes repository status check for large repositories
# much, much faster.
DISABLE_UNTRACKED_FILES_DIRTY="true"

# Uncomment the following line if you want to change the command execution time
# stamp shown in the history command output.
# You can set one of the optional three formats:
# "mm/dd/yyyy"|"dd.mm.yyyy"|"yyyy-mm-dd"
# or set a custom format using the strftime function format specifications,
# see 'man strftime' for details.
HIST_STAMPS="yyyy-mm-dd"

# Would you like to use another custom folder than $ZSH/custom?
# ZSH_CUSTOM=/path/to/new-custom-folder

# Which plugins would you like to load?
# Standard plugins can be found in $ZSH/plugins/
# Custom plugins may be added to $ZSH_CUSTOM/plugins/
# Example format: plugins=(rails git textmate ruby lighthouse)
# Add wisely, as too many plugins slow down shell startup.
plugins=(
	# common-aliases # This plugin auto-expands aliases which can be annoying
	# dirhistory - will break opt + side arrow
	# globalias # This plugin also auto-expands aliases
	# per-directory-history - really annoying
	1password
	alias-finder
	argocd
	asdf
	autoenv
	aws
	azure
	bazel
	bbedit
	bgnotify
	brew
	buf
	bun
	colored-man-pages
	colorize
	command-not-found
	cp
	dircycle
	direnv
	docker
	docker-compose
	doctl
	# dotenv
	dotnet
	emoji
	emoji-clock
	encode64
	extract
	gcloud
	gh
	git
	git-lfs
	github
	gnu-utils
	golang
	gpg-agent
	helm
	# history - just some useless aliases that conflict with hl
	history-substring-search
	jenv
	jsontools
	keychain
	kubectl
	kubectx
	macos
	mvn
	node
	nvm
	pip
	pyenv
	pylint
	python
	redis-cli
	rust
	safe-paste
	scd
	screen
	shrink-path

	ssh
	sudo
	ssh-agent
	stripe
	systemd
	tmux
	tmuxinator
	swiftpm
	terraform
	textmate
	thefuck
	urltools
	vscode
	web-search
	xcode
	zoxide
	zsh-interactive-cd
	zsh-navigation-tools

	# custom plugins
	# zsh-github-copilot # Removed as it may be causing auto-expansion
	# zsh-core
	# zsh-k9s
	# zsh-you-should-use
	zsh-fzf-history-search
	# aliasrc
)

if command -v zoxide &> /dev/null; then
	ZOXIDE_CMD_OVERRIDE="cd"
fi

# zstyle ':completion:*' rehash true

source $ZSH/oh-my-zsh.sh

# User configuration

# export MANPATH="/usr/local/man:$MANPATH"

# You may need to manually set your language environment
# export LANG=en_US.UTF-8

# Preferred editor for local and remote sessions
# if [[ -n $SSH_CONNECTION ]]; then
#   export EDITOR='vim'
# else
#   export EDITOR='nvim'
# fi

# Compilation flags
# export ARCHFLAGS="-arch $(uname -m)"

# Set personal aliases, overriding those provided by Oh My Zsh libs,
# plugins, and themes. Aliases can be placed here, though Oh My Zsh
# users are encouraged to define aliases within a top-level file in
# the $ZSH_CUSTOM folder, with .zsh extension. Examples:
# - $ZSH_CUSTOM/aliases.zsh
# - $ZSH_CUSTOM/macos.zsh
# For a full list of active aliases, run `alias`

if command -v go &> /dev/null; then
	export PATH=$(go env GOPATH)/bin:$PATH
fi

if command -v bun &> /dev/null; then
	export BUN_INSTALL="$HOME/.bun"
	export PATH="$BUN_INSTALL/bin:$PATH"
fi

if command -v jenv &> /dev/null; then
	if [ -f "$HOME/.jenv/shims/.jenv-shim" ]; then
		# if it was created greater than 5 seconds ago, delete it
		if [ "$(date +%s)" -gt "$(stat -f "%c" "$HOME/.jenv/shims/.jenv-shim")" ]; then
			echo "🔑 jenv shim is older than 5 seconds, deleting"
			rm "$HOME/.jenv/shims/.jenv-shim"
		fi
	fi

	function sync-jenv() {
		echo "🔑 syncing jenv"
		for d in /Library/Java/JavaVirtualMachines/*.jdk/Contents/Home; do jenv add "$d"; done
	}

	function reset-jenv() {
		echo "🔑 resetting jenv"
		brew uninstall jenv
		brew install jenv
		rm -rf "$HOME/.jenv"
		jenv init -
		sync-jenv
		jenv enable-plugin export
		jenv enable-plugin maven
	}
fi

# https://zsh.sourceforge.io/Doc/Release/Options.html
# https://postgresqlstan.github.io/cli/zsh-history-options
# https://unix.stackexchange.com/questions/273861/unlimited-history-in-zsh

setopt EXTENDED_HISTORY     # include timestamp
setopt BANG_HIST            # Treat the '!' character specially during expansion.
setopt HIST_BEEP            # beep if attempting to access a history entry which isn’t there
setopt HIST_FIND_NO_DUPS    # do not display previously found command
setopt HIST_IGNORE_DUPS     # do not save duplicate of prior command
setopt HIST_NO_STORE        # do not save history commands
setopt HIST_REDUCE_BLANKS   # strip superfluous blanks
setopt HIST_SAVE_NO_DUPS    # do not save duplicate entries in the history file.
setopt HIST_IGNORE_ALL_DUPS # Delete old recorded entry if new entry is a duplicate.
setopt SHARE_HISTORY        # Share history between all sessions. (conflicts with inc_append_history bc it provides this functionality)
setopt HIST_VERIFY          # Don't execute immediately upon history expansion.
setopt HIST_LEX_WORDS       # Perform history expansion on the words following the cursor, not the entire line.
setopt HIST_FCNTL_LOCK      # Use fcntl(2) to lock the history file, for thread safety.
setopt HIST_IGNORE_SPACE    # do not save if line starts with space (good way to avoid secret keys)

HISTFILE=$HOME/.config/.custom_zsh_history
SAVEHIST=999999999
HISTSIZE=999999999
