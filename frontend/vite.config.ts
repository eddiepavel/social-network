import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    allowedHosts: [
      'fipwnwbppb.a.pinggy.link',
      '.pinggy.link', // Allow all pinggy.link subdomains
    ],
  },
})
