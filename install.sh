#!/bin/sh

set -eu

if [ -x "$(command -v python3)" ]; then
  alias any_python='python3'
elif [ -x "$(command -v python)" ]; then
  alias any_python='python'
elif [ -x "$(command -v python2)" ]; then
  alias any_python='python2'
else
  echo Python 2 or 3 is required to install ias
  exit 1
fi

if [ -z "${APP_VERSION:-}" ]; then
  APP_VERSIONS=$(curl -sH"Accept: application/vnd.github.v3+json" https://api.github.com/repos/nmnellis/istio-app-simulator/releases | any_python -c "import sys; from distutils.version import StrictVersion, LooseVersion; from json import loads as l; releases = l(sys.stdin.read()); releases = [release['tag_name'] for release in releases];  filtered_releases = list(filter(lambda release_string: len(release_string) > 0 and StrictVersion.version_re.match(release_string[1:]) != None, releases)); filtered_releases.sort(key=LooseVersion, reverse=True); print('\n'.join(filtered_releases))")
else
  APP_VERSIONS="${APP_VERSION}"
fi

if [ "$(uname -s)" = "Darwin" ]; then
  OS=darwin
else
  OS=linux
fi

for app_version in $APP_VERSIONS; do

tmp=$(mktemp -d /tmp/ias.XXXXXX)
filename="istio_app_simulator_${OS}_amd64"
url="https://github.com/nmnellis/istio-app-simulator/releases/download/${app_version}/${filename}.zip"

if curl -f ${url} >/dev/null 2>&1; then
  echo "Attempting to download istio app simulator version ${app_version}"
else
  continue
fi

(
  cd "$tmp"

  echo "Downloading ${filename}..."

  SHA=$(curl -sL "${url}.sha256" | cut -d' ' -f1)
  curl -sLO "${url}"
  echo "Download complete!"
#  checksum=$(openssl dgst -sha256 "${filename}" | awk '{ print $2 }')
#  if [ "$checksum" != "$SHA" ]; then
#    echo "Checksum validation failed." >&2
#    exit 1
#  fi
#  echo "Checksum valid."

)

(

  cd "$HOME"
  mkdir -p ".istio/bin"
  echo "extracting zip ${tmp}/${filename}"
  unzip "${tmp}/${filename}.zip" -d ${tmp}
  mv "${tmp}/istio-app-simulator" ".istio/bin/ias"
  chmod +x ".istio/bin/ias"
)

rm -r "$tmp"

echo "istio app simulator was successfully installed 🎉"
echo ""
echo "Add the ias CLI to your path with:"
echo "  export PATH=\$HOME/.istio/bin:\$PATH"
echo ""
echo "Now run:"
echo "  ias generate     # generate applications"
exit 0
done

echo "No versions of ias found."
exit 1
