#!/bin/sh

# The install script is licensed under the MIT license Glide itself is under.
# See https://github.com/Masterminds/glide/blob/master/LICENSE for more details.

# To run this script execute:
#   `curl https://glide.sh/get | sh`

PROJECT_NAME="glide"

# LGOBIN represents the local bin location. This can be either the GOBIN, if set,
# or the GOPATH/bin.

LGOBIN=""

fail() {
	echo "$1"
	exit 1
}

verifyGoInstallation() {
	GO=$(which go)
	if [ "$?" = "1" ]; then
		fail "$PROJECT_NAME needs go. Please install it first."
	fi
	if [ -z "$GOPATH" ]; then
	    # Verify if default GOPATH of Go >=1.8 is used
	    GOPATH=$(go env GOPATH)
	    if [ ! -d "$GOPATH" ]; then
		fail "$PROJECT_NAME needs environment variable "'$GOPATH'". Set it before continue."
	    fi
	fi
	if [ -n "$GOBIN" ]; then
		if [ ! -d "$GOBIN" ]; then
			fail "$GOBIN "'($GOBIN)'" folder not found. Please create it before continue."
		fi
		LGOBIN="$GOBIN"
	else
		if [ ! -d "$GOPATH/bin" ]; then
			fail "$GOPATH/bin "'($GOPATH/bin)'" folder not found. Please create it before continue."
		fi
		LGOBIN="$GOPATH/bin"
	fi

}

initArch() {
	ARCH=$(uname -m)
	case $ARCH in
		armv5*) ARCH="armv5";;
		armv6*) ARCH="armv6";;
		armv7*) ARCH="armv7";;
		aarch64) ARCH="arm64";;
		x86) ARCH="386";;
		x86_64) ARCH="amd64";;
		i686) ARCH="386";;
		i386) ARCH="386";;
	esac
	echo "ARCH=$ARCH"
}

initOS() {
	OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

	case "$OS" in
		# Minimalist GNU for Windows
		mingw*) OS='windows';;
		msys*) OS='windows';;
	esac
	echo "OS=$OS"
}

initDownloadTool() {
	if type "curl" > /dev/null; then
		DOWNLOAD_TOOL="curl"
	elif type "wget" > /dev/null; then
		DOWNLOAD_TOOL="wget"
	else
		fail "You need curl or wget as download tool. Please install it first before continue"
	fi
	echo "Using $DOWNLOAD_TOOL as download tool"
}

get() {
	local url="$2"
	local body
	local httpStatusCode
	echo "Getting $url"
	if [ "$DOWNLOAD_TOOL" = "curl" ]; then
		httpResponse=$(curl -sL --write-out HTTPSTATUS:%{http_code} "$url")
		httpStatusCode=$(echo $httpResponse | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
		body=$(echo "$httpResponse" | sed -e 's/HTTPSTATUS\:.*//g')
	elif [ "$DOWNLOAD_TOOL" = "wget" ]; then
		tmpFile=$(mktemp)
		body=$(wget --server-response --content-on-error -q -O - "$url" 2> $tmpFile || true)
		httpStatusCode=$(cat $tmpFile | awk '/^  HTTP/{print $2}')
	fi
	if [ "$httpStatusCode" != 200 ]; then
		echo "Request fail with http status code $httpStatusCode"
		fail "Body: $body"
	fi
	eval "$1='$body'"
}

getFile() {
	local url="$1"
	local filePath="$2"
	if [ "$DOWNLOAD_TOOL" = "curl" ]; then
		httpStatusCode=$(curl -s -w '%{http_code}' -L "$url" -o "$filePath")
	elif [ "$DOWNLOAD_TOOL" = "wget" ]; then
		body=$(wget --server-response --content-on-error -q -O "$filePath" "$url")
		httpStatusCode=$(cat $tmpFile | awk '/^  HTTP/{print $2}')
	fi
	echo "$httpStatusCode"
}


downloadFile() {
	# get TAG https://glide.sh/version
	export TAG=v0.13.1
	echo "TAG=$TAG"
	GLIDE_DIST="glide-$TAG-$OS-$ARCH.tar.gz"
	echo "GLIDE_DIST=$GLIDE_DIST"
	DOWNLOAD_URL="https://github.com/Masterminds/$PROJECT_NAME/releases/download/$TAG/$GLIDE_DIST"
	GLIDE_TMP_FILE="/tmp/$GLIDE_DIST"
	echo "Downloading $DOWNLOAD_URL"
	httpStatusCode=$(getFile "$DOWNLOAD_URL" "$GLIDE_TMP_FILE")
	if [ "$httpStatusCode" -ne 200 ]; then
		echo "Did not find a release for your system: $OS $ARCH"
		echo "Trying to find a release on the github api."
		LATEST_RELEASE_URL="https://api.github.com/repos/Masterminds/$PROJECT_NAME/releases/tags/$TAG"
		echo "LATEST_RELEASE_URL=$LATEST_RELEASE_URL"
		get LATEST_RELEASE_JSON $LATEST_RELEASE_URL
		# || true forces this command to not catch error if grep does not find anything
		DOWNLOAD_URL=$(echo "$LATEST_RELEASE_JSON" | grep 'browser_' | cut -d\" -f4 | grep "$GLIDE_DIST") || true
		if [ -z "$DOWNLOAD_URL" ]; then
			echo "Sorry, we dont have a dist for your system: $OS $ARCH"
			fail "You can ask one here: https://github.com/Masterminds/$PROJECT_NAME/issues"
		else
			echo "Downloading $DOWNLOAD_URL"
			getFile "$DOWNLOAD_URL" "$GLIDE_TMP_FILE"
		fi
	fi
}

installFile() {
	GLIDE_TMP="/tmp/$PROJECT_NAME"
	mkdir -p "$GLIDE_TMP"
	tar xf "$GLIDE_TMP_FILE" -C "$GLIDE_TMP"
	GLIDE_TMP_BIN="$GLIDE_TMP/$OS-$ARCH/$PROJECT_NAME"
	cp "$GLIDE_TMP_BIN" "$LGOBIN"
	rm -rf $GLIDE_TMP
	rm -f $GLIDE_TMP_FILE
}

bye() {
	result=$?
	if [ "$result" != "0" ]; then
		echo "Fail to install $PROJECT_NAME"
	fi
	exit $result
}

testVersion() {
	set +e
	GLIDE="$(which $PROJECT_NAME)"
	if [ "$?" = "1" ]; then
		fail "$PROJECT_NAME not found. Did you add "'$GOBIN'" to your "'$PATH?'
	fi
	set -e
	GLIDE_VERSION=$($PROJECT_NAME -v)
	echo "$GLIDE_VERSION installed successfully"
}


# Execution

#Stop execution on any error
trap "bye" EXIT
verifyGoInstallation
set -e
initArch
initOS
initDownloadTool
downloadFile
installFile
testVersion
