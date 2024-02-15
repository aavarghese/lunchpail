import type Kind from "@jaas/common/Kind"

import { Dashboard } from "../pages/Dashboard"

import DemoRunEventSource from "./streams/run"
import NothingEventSource from "./streams/nothing"
import DemoQueueEventSource from "./streams/queue"
import DemoTaskQueueEventSource from "./streams/taskqueue"
import DemoWorkerPoolStatusEventSource from "./streams/pool"
import DemoApplicationSpecEventSource from "./streams/application"

import context from "./context"

import type WatchedKind from "@jaas/common/Kind"
import type DemoEventSource from "./streams/base"
import type KubernetesResource from "@jaas/common/events/KubernetesResource"

let props: null | (Record<WatchedKind, DemoEventSource> & { workerpools: DemoWorkerPoolStatusEventSource }) = null

function init() {
  if (props === null) {
    const runs = new DemoRunEventSource()
    const queues = new DemoQueueEventSource()
    const taskqueues = new DemoTaskQueueEventSource()
    const workerpools = new DemoWorkerPoolStatusEventSource(taskqueues, queues)
    const applications = new DemoApplicationSpecEventSource()

    props = {
      runs,
      computetargets: new NothingEventSource(),
      taskqueues,
      datasets: new NothingEventSource(),
      workerpools,
      queues,
      applications,
      platformreposecrets: new NothingEventSource(),
      workdispatchers: new NothingEventSource(),
    }
  }

  return props
}

export default function DemoDashboard() {
  const props = init()

  if (!window.demo) {
    window.jaas = window.demo = Object.assign({}, props, {
      create: props.workerpools.create.bind(props.workerpools),

      async delete(yaml: string) {
        const { loadAll } = await import("js-yaml")
        const rsrcs = loadAll(yaml) as KubernetesResource[]
        await Promise.all(
          rsrcs.map((rsrc) =>
            window.demo.deleteByName({
              kind: (rsrc.kind.toLowerCase() + "s") as Kind,
              name: rsrc.metadata.name,
              namespace: rsrc.metadata.namespace,
              context,
            }),
          ),
        )
        return true as const
      },

      /** Delete the given named `ComputeTarget` */
      deleteComputeTarget() {
        throw new Error("Unsupported operation")
      },

      deleteByName(dprops: import("@jaas/common/api/jaas").DeleteProps) {
        if (/workerpool/.test(dprops.kind)) {
          return props.workerpools.delete(dprops)
        } else if (/run/.test(dprops.kind)) {
          return props.runs.delete(dprops)
        } else if (/application/.test(dprops.kind)) {
          return props.applications.delete(dprops)
        } else if (/taskqueue/.test(dprops.kind)) {
          return props.taskqueues.delete(dprops)
        } else {
          return {
            code: 404,
            message: "Resource not found",
          }
        }
      },
      controlplane: {
        async status(): Promise<import("@jaas/common/status/ControlPlaneStatus").default> {
          return {
            containerCLI: "podman" as const,
            containerRuntime: "podman" as const,
            containerRuntimeOnline: true,
            kubernetesClusterExists: true,
            kubernetesClusterOnline: true,
          }
        },
        init() {},
        update() {},
        destroy() {},
      },
    })
  }

  return <Dashboard {...props} />
}
