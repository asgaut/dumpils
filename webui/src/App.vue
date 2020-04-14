<template>
  <div id="app">
    <div id="nav">
      <router-link to="/">Home</router-link>&nbsp;|
      <router-link to="/spectrum">Spectrum</router-link>&nbsp;|
      <router-link to="/generator">Generator</router-link>&nbsp;|
      <router-link to="/about">About</router-link>
    </div>
    <router-view />
  </div>
</template>

<style>
/* https://github.com/microsoft/vscode/extensions/theme-defaults/themes/dark_defaults.json */

body {
  background-color: #1e1e1e;
}

#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #bbbbbb;
}

#nav {
  text-align: center;
  padding: 30px;
  background: #252526;
}

#nav a {
  font-weight: bold;
  color: #cccccc;
}

a {
  color: #d4d4d4;
}

a:visited {
  color: #bbbbbb;
}
</style>
