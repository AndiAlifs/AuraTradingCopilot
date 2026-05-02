/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts}",
  ],
  theme: {
    extend: {
      colors: {
        auraGreen: '#4ade80',
        auraRed: '#f87171',
        auraDark: '#0f172a', /* slate-900 */
        auraPanel: '#1e293b', /* slate-800 */
      }
    },
  },
  plugins: [],
}
