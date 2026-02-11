/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Neo Brutalism - Bold Orange primary
        primary: {
          50: '#fff7ed',
          100: '#ffedd5',
          200: '#fed7aa',
          300: '#fdba74',
          400: '#fb923c',
          500: '#FF6B00',
          600: '#ea580c',
          700: '#c2410c',
          800: '#9a3412',
          900: '#7c2d12',
          950: '#431407'
        },
        // Neo Brutalism - Neutrals
        accent: {
          50: '#fafaf9',
          100: '#f5f5f4',
          200: '#e7e5e4',
          300: '#d6d3d1',
          400: '#a8a29e',
          500: '#78716c',
          600: '#57534e',
          700: '#44403c',
          800: '#292524',
          900: '#1c1917',
          950: '#0c0a09'
        },
        // Dark mode backgrounds
        dark: {
          50: '#fafaf9',
          100: '#f5f5f4',
          200: '#e7e5e4',
          300: '#d6d3d1',
          400: '#a8a29e',
          500: '#78716c',
          600: '#57534e',
          700: '#44403c',
          800: '#292524',
          900: '#1c1917',
          950: '#0c0a09'
        },
        // Neo Brutalism surface colors
        cream: '#FFF8F0',
        brutal: {
          black: '#1A1A1A',
          white: '#FFFFFF',
          orange: '#FF6B00',
          teal: '#00C9A7',
          cream: '#FFF8F0',
          yellow: '#FFD600',
          pink: '#FF5C8A',
          blue: '#3B82F6',
        }
      },
      fontFamily: {
        sans: [
          'Space Grotesk',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['Space Mono', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        // Neo Brutalism hard shadows
        'brutal': '4px 4px 0px #1A1A1A',
        'brutal-sm': '2px 2px 0px #1A1A1A',
        'brutal-md': '4px 4px 0px #1A1A1A',
        'brutal-lg': '6px 6px 0px #1A1A1A',
        'brutal-xl': '8px 8px 0px #1A1A1A',
        'brutal-hover': '6px 6px 0px #1A1A1A',
        'brutal-active': '2px 2px 0px #1A1A1A',
        'brutal-orange': '4px 4px 0px #FF6B00',
        'brutal-teal': '4px 4px 0px #00C9A7',
        // Keep some for compatibility
        glass: '4px 4px 0px #1A1A1A',
        'glass-sm': '2px 2px 0px #1A1A1A',
        glow: '4px 4px 0px #FF6B00',
        'glow-lg': '6px 6px 0px #FF6B00',
        card: '4px 4px 0px #1A1A1A',
        'card-hover': '6px 6px 0px #1A1A1A',
        'inner-glow': 'none'
      },
      backgroundImage: {
        // Remove all gradients for Neo Brutalism - flat colors only
        'gradient-radial': 'none',
        'gradient-primary': 'none',
        'gradient-dark': 'none',
        'gradient-glass': 'none',
        'mesh-gradient': 'none'
      },
      animation: {
        'fade-in': 'fadeIn 0.2s ease-out',
        'slide-up': 'slideUp 0.2s ease-out',
        'slide-down': 'slideDown 0.2s ease-out',
        'slide-in-right': 'slideInRight 0.2s ease-out',
        'scale-in': 'scaleIn 0.15s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
      },
      borderWidth: {
        '3': '3px',
      },
      borderRadius: {
        'none': '0',
        'brutal': '0',
      }
    }
  },
  plugins: []
}
