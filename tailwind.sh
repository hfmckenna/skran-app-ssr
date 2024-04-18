curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
chmod +x tailwindcss-macos-arm64
mv tailwindcss-macos-arm64 tailwindcss
./tailwindcss -i index.css -o public/index.css --minify
aws s3 cp ./public s3://skran-app-ssr-assets --profile sso-dev --recursive