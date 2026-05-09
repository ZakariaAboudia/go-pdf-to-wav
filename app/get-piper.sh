#!/bin/bash
set -e

wget https://github.com/rhasspy/piper/releases/download/2023.11.14-2/piper_linux_x86_64.tar.gz \
  && tar -xvzf piper_linux_x86_64.tar.gz \
  && rm piper_linux_x86_64.tar.gz

echo "Downloading piper default voice"

mkdir -p ./piper/voices
wget -q -O ./piper/voices/en_US-libritts_r-medium.onnx \
  https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/libritts_r/medium/en_US-libritts_r-medium.onnx
wget -q -O ./piper/voices/en_US-libritts_r-medium.onnx.json \
  https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/libritts_r/medium/en_US-libritts_r-medium.onnx.json
