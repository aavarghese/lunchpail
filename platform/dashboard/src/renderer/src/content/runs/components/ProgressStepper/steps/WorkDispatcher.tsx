import { linkToAllDetails } from "@jaas/renderer/navigate/details"

import taskqueueProps from "@jaas/resources/runs/components/taskqueueProps"
import NewWorkDispatcherButton from "@jaas/resources/runs/components/actions/NewWorkDispatcher"

import type Step from "../Step"
import { oopsNoQueue } from "../oops"
import workdispatchers from "@jaas/resources/runs/components/workdispatchers"

import { status } from "@jaas/resources/workdispatchers/status"
import { singular as run } from "@jaas/resources/runs/name"
import { name as workerpools } from "@jaas/resources/workerpools/name"
import { singular as taskqueue } from "@jaas/resources/taskqueues/name"
import { group as dispatch } from "@jaas/resources/workdispatchers/group"
import { singular as workdispatcher } from "@jaas/resources/workdispatchers/name"

import type Props from "@jaas/resources/runs/components/Props"

function statusOfDispatchers(props: Props) {
  const all = workdispatchers(props)

  const pending = all.filter((_) => /Pend/i.test(status(_))).length
  const running = all.filter((_) => /Run/i.test(status(_))).length
  const finished = all.filter((_) => /Succe/i.test(status(_))).length
  const failed = all.filter((_) => /Fail/i.test(status(_))).length

  return { pending, running, finished, failed }
}

function variant(props: Props) {
  const { pending, running, finished, failed } = statusOfDispatchers(props)

  return pending + running + finished + failed === 0
    ? ("warning" as const)
    : failed > 0
      ? ("danger" as const)
      : pending > 0
        ? ("pending" as const)
        : running > 0
          ? ("info" as const)
          : ("success" as const)
}

const step: Step = {
  id: dispatch,
  variant,
  content: (props, onClick) => {
    const queue = taskqueueProps(props)
    const dispatchers = workdispatchers(props)

    if (!queue) {
      return oopsNoQueue
    } else if (dispatchers.length === 0) {
      const body = (
        <span>
          A <strong>{workdispatcher}</strong> will populate the {taskqueue} for a {run}. Any assigned{" "}
          <strong>{workerpools}</strong> will then consume these Tasks.{" "}
        </span>
      )

      const footer = <NewWorkDispatcherButton isInline {...props} queueProps={queue} onClick={onClick} />

      return { body, footer }
    } else {
      return linkToAllDetails("workdispatchers", dispatchers)
    }
  },
}

export default step