// Bindings for the Go App methods (normally made by `wails generate`).
export function Connect() {
  return window["go"]["main"]["App"]["Connect"]();
}

export function Devices() {
  return window["go"]["main"]["App"]["Devices"]();
}

export function Disconnect() {
  return window["go"]["main"]["App"]["Disconnect"]();
}

export function Hide() {
  return window["go"]["main"]["App"]["Hide"]();
}

export function HotspotStart() {
  return window["go"]["main"]["App"]["HotspotStart"]();
}

export function HotspotStatus() {
  return window["go"]["main"]["App"]["HotspotStatus"]();
}

export function HotspotStop() {
  return window["go"]["main"]["App"]["HotspotStop"]();
}

export function Minimise() {
  return window["go"]["main"]["App"]["Minimise"]();
}

export function OpenLogFolder() {
  return window["go"]["main"]["App"]["OpenLogFolder"]();
}

export function OpenURL(arg1) {
  return window["go"]["main"]["App"]["OpenURL"](arg1);
}

export function Quit() {
  return window["go"]["main"]["App"]["Quit"]();
}

export function SaveSettings(arg1) {
  return window["go"]["main"]["App"]["SaveSettings"](arg1);
}

export function SetAutostart(arg1) {
  return window["go"]["main"]["App"]["SetAutostart"](arg1);
}

export function Settings() {
  return window["go"]["main"]["App"]["Settings"]();
}

export function Show() {
  return window["go"]["main"]["App"]["Show"]();
}

export function Snapshot() {
  return window["go"]["main"]["App"]["Snapshot"]();
}

export function Toggle() {
  return window["go"]["main"]["App"]["Toggle"]();
}

