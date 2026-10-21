// Package assets bundles the helper files the app unpacks at run time:
// Google's adb (USB bridge) and the phone-side server for each CPU type.
package assets

import "embed"

//go:embed adb/adb.exe adb/AdbWinApi.dll adb/AdbWinUsbApi.dll phone/*
var FS embed.FS
