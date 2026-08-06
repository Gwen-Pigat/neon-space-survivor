package main

import (
	"embed"
)

//go:embed static/images/*.png static/audio/global/theme/*.mp3 static/audio/global/home/*.mp3 static/audio/global/segment_last/*.mp3 static/audio/global/alarms/*.mp3 static/audio/global/losses/*.mp3 static/audio/global/success/*.mp3 static/audio/enemies/boss/*.mp3
var assetsFS embed.FS
