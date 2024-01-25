import ExecResponse from "@jay/common/events/ExecResponse"
import { clusterNameForKubeconfig } from "../controlplane/kind"

import { isComputeTarget } from "../streams/computetargets"

/**
 * Create a resource using the given `yaml` spec.
 */
export async function onCreate(
  yaml: string,
  action: "apply" | "delete",
  context = clusterNameForKubeconfig,
  dryRun = false,
): Promise<ExecResponse> {
  const [{ spawn }, { load }] = await Promise.all([import("node:child_process"), import("js-yaml")])

  const json = load(yaml)
  if (isComputeTarget(json)) {
    return import("../streams/computetargets").then(({ createComputeTarget }) => createComputeTarget(json, dryRun))
  }

  return new Promise((resolve) => {
    try {
      // the `-f -` means accept the yaml on stdin
      const child = spawn(
        "kubectl",
        [action, "--context", context, "-f", "-", ...(dryRun === false ? [] : ["--dry-run=server"])],
        {
          stdio: ["pipe", "inherit", "pipe"],
        },
      )

      // send the yaml to the kubectl apply across stdin
      child.stdin.write(yaml)
      child.stdin.end()

      let err = ""
      child.stderr.on("data", (data) => (err += data.toString()))

      child.once("close", (code) => {
        if (code === 0) {
          resolve(true)
        } else {
          resolve({ code, message: err })
        }
      })
    } catch (err) {
      console.error(err)
      resolve({ code: 1, message: hasMessage(err) ? err.message : "" })
    }
  })
}

export function hasMessage(obj: unknown): obj is { message: string } {
  return typeof (obj as { message: string }).message === "string"
}

/**
 * Delete a resource by name
 */
export async function onDelete(yaml: string, context?: string): Promise<ExecResponse> {
  return onCreate(yaml, "delete", context, false)
}

/**
 * Delete a resource by name
 */
export async function onDeleteByName({
  kind,
  name,
  namespace,
}: import("@jay/common/api/jay").DeleteProps): Promise<ExecResponse> {
  const { spawn } = await import("node:child_process")
  return new Promise((resolve) => {
    try {
      // the `-f -` means accept the yaml on stdin
      const child = spawn("kubectl", ["delete", kind, "--context", clusterNameForKubeconfig, name, "-n", namespace], {
        stdio: ["inherit", "inherit", "pipe"],
      })

      let err = ""
      child.stderr.on("data", (data) => (err += data.toString()))

      child.once("close", (code) => {
        if (code === 0) {
          resolve(true)
        } else {
          resolve({ code, message: err })
        }
      })
    } catch (err) {
      console.error(err)
      resolve({ code: 1, message: hasMessage(err) ? err.message : "" })
    }
  })
}
