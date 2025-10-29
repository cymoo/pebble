import React from 'react'
import ReactDOM from 'react-dom/client'
import { Toaster } from 'react-hot-toast'
import { BrowserRouter } from 'react-router'

import { IS_TOUCH_DEVICE } from '@/utils/browser'
import { calcViewportUnit } from '@/utils/compt.ts'

import { ConfirmProvider } from '@/components/confirm'
import { ModalProvider } from '@/components/modal.tsx'
import { BackgroundLocationProvider, StableNavigateProvider } from '@/components/router.tsx'
import { LangProvider } from '@/components/translation'

import './index.css'
import reportWebVitals from './reportWebVitals.ts'
import { App } from './route'

const initVConsoleWhenInDevelopment = () => {
  if (import.meta.env.DEV && IS_TOUCH_DEVICE) {
    void import('vconsole').then((mod) => {
      new mod.default({ theme: 'dark' })
    })
  }
}

const initBaseFontSize = () => {
  document.documentElement.style.fontSize = localStorage.getItem('baseFontSize') || '16px'
}

const initLang = () => {
  const lang = localStorage.getItem('lang') || 'zh'
  document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en'
}

calcViewportUnit()
initBaseFontSize()
initLang()
initVConsoleWhenInDevelopment()

const root = ReactDOM.createRoot(document.getElementById('root')!)
root.render(
  <React.StrictMode>
    <LangProvider>
      <BrowserRouter basename={import.meta.env.VITE_MEMO_URL}>
        <StableNavigateProvider>
          <BackgroundLocationProvider>
            <ModalProvider>
              <ConfirmProvider>
                <App />
              </ConfirmProvider>
            </ModalProvider>
          </BackgroundLocationProvider>
        </StableNavigateProvider>
      </BrowserRouter>
      <Toaster
        toastOptions={{
          style: {
            border: '1px solid hsl(var(--border))',
            background: 'hsl(var(--popover))',
            color: 'hsl(var(--popover-foreground))',
          },
        }}
      />
    </LangProvider>
  </React.StrictMode>,
)

// To start measuring performance, pass a function to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
// https://github.com/GoogleChrome/web-vitals
reportWebVitals()
