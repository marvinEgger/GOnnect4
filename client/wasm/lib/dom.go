// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

package lib

import "syscall/js"

// GetElement returns a DOM element by ID
func GetElement(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}

// SetText sets text content of an element
func SetText(id, text string) {
	el := GetElement(id)
	if !el.IsNull() {
		el.Set("textContent", text)
	}
}

// SetStyle sets a CSS style property
func SetStyle(id, property, value string) {
	el := GetElement(id)
	if !el.IsNull() {
		el.Get("style").Set(property, value)
	}
}

// SetValue sets value of an input element
func SetValue(id, value string) {
	el := GetElement(id)
	if !el.IsNull() {
		el.Set("value", value)
	}
}

// GetValue gets value of an input element
func GetValue(id string) string {
	el := GetElement(id)
	if el.IsNull() {
		return ""
	}
	return el.Get("value").String()
}

// AddClass adds a CSS class to an element
func AddClass(id, className string) {
	el := GetElement(id)
	if !el.IsNull() {
		el.Get("classList").Call("add", className)
	}
}

// RemoveClass removes a CSS class from an element
func RemoveClass(id, className string) {
	el := GetElement(id)
	if !el.IsNull() {
		el.Get("classList").Call("remove", className)
	}
}

// ToggleClass toggles a CSS class on an element
func ToggleClass(id, className string, force bool) {
	el := GetElement(id)
	if !el.IsNull() {
		el.Get("classList").Call("toggle", className, force)
	}
}

// SetDisplay sets display style property
func SetDisplay(id, value string) {
	SetStyle(id, "display", value)
}

// Show shows an element
func Show(id string) {
	SetDisplay(id, "block")
}

// ShowFlex shows an element with flex display
func ShowFlex(id string) {
	SetDisplay(id, "flex")
}

// Hide hides an element
func Hide(id string) {
	SetDisplay(id, "none")
}

// ShowScreen shows a specific screen
func ShowScreen(name string) {
	screens := []string{"login-screen", "lobby-screen", "game-screen"}

	for _, screen := range screens {
		RemoveClass(screen, "active")
	}

	AddClass(name+"-screen", "active")
}

// ShowMessage displays a message
func ShowMessage(elementID, text, msgType string) {
	el := GetElement(elementID)
	if el.IsNull() {
		return
	}

	el.Set("textContent", text)
	el.Set("className", "message "+msgType)
}

// SetLocalStorage sets an item in localStorage
func SetLocalStorage(key, value string) {
	js.Global().Get("localStorage").Call("setItem", key, value)
}

// GetLocalStorage gets an item from localStorage
func GetLocalStorage(key string) string {
	val := js.Global().Get("localStorage").Call("getItem", key)
	if val.IsNull() {
		return ""
	}
	return val.String()
}

// RemoveLocalStorage removes an item from localStorage
func RemoveLocalStorage(key string) {
	js.Global().Get("localStorage").Call("removeItem", key)
}

// Alert shows a browser alert
func Alert(message string) {
	js.Global().Call("alert", message)
}

// Confirm shows a confirmation dialog
func Confirm(message string) bool {
	return js.Global().Call("confirm", message).Bool()
}

// Console logs to browser console
func Console(message string) {
	js.Global().Get("console").Call("log", message)
}
