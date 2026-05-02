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
        auraDark: '#020617', /* slate-950 deep slate background */
        auraPanel: '#0f172a', /* slate-900 */
        auraBorder: '#1e293b', /* slate-800 */
        auraNeon: '#06b6d4', /* cyan-500 neon accent */
      },
      boxShadow: {
        'neon': '0 0 10px rgba(6, 182, 212, 0.5), 0 0 20px rgba(6, 182, 212, 0.3)',
      }
    },
  },
  plugins: [],
}
