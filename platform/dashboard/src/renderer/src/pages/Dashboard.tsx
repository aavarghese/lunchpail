import { useContext, useEffect, lazy, Suspense } from "react"

const Modal = lazy(() => import("@patternfly/react-core").then((_) => ({ default: _.Modal })))
const MissingWizardError = lazy(() => import("../components/MissingWizardError"))

import { currentKind } from "../navigate/kind"
import { isShowingWizard } from "../navigate/wizard"
import { returnHomeCallback } from "../navigate/home"

import drilldownProps from "./DrilldownProps"
import PageWithDrawer from "./PageWithDrawer"

import Settings from "../Settings"
import Sidebar from "../components/Sidebar"

import content from "../content/providers"
import { initMemos } from "../content/memos"
import initStreamingState from "../content/init"

import type WatchedKind from "@jaas/common/Kind"
import type EventSourceLike from "@jaas/common/events/EventSourceLike"

import uniqueTaskQueues from "@jaas/resources/taskqueues/unique"

import "./Dashboard.scss"

/** one EventSource per resource Kind */
export type Props<Source extends EventSourceLike = EventSourceLike> = Record<WatchedKind, Source>

export function Dashboard(props: Props) {
  const settings = useContext(Settings)
  const inDemoMode = settings?.demoMode[0] ?? false

  const returnHome = returnHomeCallback()

  const { events, handlers } = initStreamingState()
  const memos = initMemos(events)

  // This registers what is in effect a componentDidMount handler. We
  // use it to register/deregister our event `handlers`
  useEffect(function onMount() {
    Object.entries(handlers).forEach(([kind, handler]) => {
      props[kind].addEventListener("message", handler, false)
    })

    // return a cleanup function to be called when the component unmounts
    return () =>
      Object.entries(handlers).forEach(([kind, handler]) => props[kind].removeEventListener("message", handler))
  }, [])

  /** Content to display in the slide-out Drawer panel */
  const { currentlySelectedId: id, currentlySelectedKind: kind, currentlySelectedContext: context } = drilldownProps()
  const detailContentProvider = id && kind && content[kind]
  const currentDetail =
    detailContentProvider && detailContentProvider.detail
      ? detailContentProvider.detail(id, context || "", events, memos, settings)
      : undefined
  const panelSubtitle = currentDetail ? currentDetail.subtitle : undefined
  const panelBody = currentDetail ? currentDetail.body : undefined

  /** Content to display in the main gallery */
  const bodyContentProvider = content[currentKind()]
  const currentActions =
    bodyContentProvider && bodyContentProvider.actions ? bodyContentProvider.actions({ inDemoMode }) : undefined

  /** Content to display in the modal */
  const kindForWizard = isShowingWizard()
  const wizardContentProvider = !!kindForWizard && content[kindForWizard]
  const missingWizard = wizardContentProvider && !wizardContentProvider.wizard
  const modal = (
    <Suspense fallback={<></>}>
      <Modal
        variant="large"
        showClose={missingWizard /* MissingWizardError needs us to provide a close button */}
        hasNoBodyWrapper={!missingWizard}
        aria-label="wizard-modal"
        onEscapePress={returnHome}
        onClose={returnHome}
        isOpen={!!wizardContentProvider}
      >
        {wizardContentProvider && wizardContentProvider.wizard ? (
          wizardContentProvider.wizard(events)
        ) : kindForWizard ? (
          <MissingWizardError kind={kindForWizard} />
        ) : undefined}
      </Modal>
    </Suspense>
  )

  /** Content to display in the hamburger-menu sidebar (usually coming in on the left) */
  const sidebar = (
    <Sidebar
      runs={events.runs.length}
      taskqueues={uniqueTaskQueues(events).length}
      datasets={events.datasets.length}
      workerpools={events.workerpools.length}
      applications={events.applications.length}
      workdispatchers={events.workdispatchers.length}
      platformreposecrets={events.platformreposecrets.length}
      computetargets={events.computetargets.filter((_) => _.spec.isJaaSWorkerHost).length}
    />
  )

  const pwdProps = {
    panelSubtitle,
    panelBody,
    modal,
    title: bodyContentProvider.title ?? bodyContentProvider.name,
    subtitle: bodyContentProvider.description,
    sidebar,
    actions: currentActions,
  }

  return (
    <PageWithDrawer {...pwdProps}>
      {bodyContentProvider.gallery && bodyContentProvider.gallery(events, memos, settings)}
    </PageWithDrawer>
  )
}
