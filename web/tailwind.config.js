// tailwind.config.js
/** @type {import('tailwindcss').Config} */
const defaultTheme = require('tailwindcss/defaultTheme')
const colors = require('tailwindcss/colors')
module.exports = {
  content: [
    "./static/js/*.js",
    "./static/images/*.svg",
    "./static/icons/*.svg",
    "./templates/**/*.html",
    "./templates/**/*.css",
    "./lib/*.ts",
    "./lib/*.tsx",
    "./src/*.ts",
    "./src/*.tsx"
  ],
  darkMode: 'class',
  theme: {
  },
  plugins: [require('@tailwindcss/forms')/*, require('@tailwindcss/typography')*/],
  extend: {
      screens: {
      print: { raw: 'print' },
      screen: { raw: 'screen' },
    },
  }
}
