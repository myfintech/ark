#!/bin/bash
set -e

install_on_linux () {
  curl --fail --fail-early -Lo ark "{{downloadBaseURL}}/arkcliv2-linux"

  if [[ "$*" == *--docker* ]]
  then
    install ark "/usr/local/bin"
    mkdir -p /var/log/ark
    chmod a+w /var/log/ark
  else
    mkdir -p "$HOME/.local/bin/"
    sudo install ark "$HOME/.local/bin/"
    sudo mkdir -p /var/log/ark
    sudo chmod a+w /var/log/ark
  fi

  rm ark
}

install_on_darwin () {
  if [[ $( id -u ) == "0" ]]
  then
      echo "please DO NOT run this script with sudo"
      exit 1
  fi
  curl --fail --fail-early -Lo ark "{{downloadBaseURL}}/arkcliv2-darwin"
  install ark /usr/local/bin/
  rm ark
}

main() {
    # escaping is required to use docker heredoc syntax
    case "$(uname -s).$(uname -m)" in
        Linux.x86_64) install_on_linux "$@";;
        Darwin.x86_64) install_on_darwin "$@";;
        *) echo "sorry, there is no binary distribution of Ark for your platform"; exit 1;;
    esac
}

main "$@"