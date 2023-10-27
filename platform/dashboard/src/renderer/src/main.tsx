import { createRoot } from "react-dom/client"
import { RouterProvider } from "react-router-dom"

import router from "./router"
import Status, { statusState } from "./Status"
import Settings, { darkModeState, demoModeState, formState } from "./Settings"

import "@patternfly/react-core/dist/styles/base.css"

function App() {
  const darkMode = darkModeState() // UI in dark mode?
  const demoMode = demoModeState() // are we running in offline mode?
  const form = formState() // remember form choices made in wizards

  // is the local control plane good to go? null means unknown
  // (e.g. that a check is in progress)
  const statusCtx = statusState(demoMode)

  return (
    <Settings.Provider value={{ darkMode, demoMode, form }}>
      <Status.Provider value={statusCtx}>
        <RouterProvider router={router} />
      </Status.Provider>
    </Settings.Provider>
  )
}

createRoot(document.getElementById("root") as HTMLElement).render(<App />)
