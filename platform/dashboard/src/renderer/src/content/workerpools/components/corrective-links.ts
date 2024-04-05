import type Props from "./Props"
import { buttonPropsForNewRepoSecret } from "@jaas/renderer/navigate/newreposecret"

/** Any suggestions/corrective action links */
export default function correctiveLinks(
  location: Parameters<typeof buttonPropsForNewRepoSecret>[0],
  props: Props,
  startOrAdd: "fix" | "create" = "fix",
) {
  const latestStatus = props.status
  const status = latestStatus?.metadata.annotations["lunchpail.io/status"]
  const reason = latestStatus?.metadata.annotations["lunchpail.io/reason"]
  const message = latestStatus?.metadata.annotations["lunchpail.io/message"]
  if (status === "CloneFailed" && reason === "AccessDenied") {
    const repoMatch = message?.match(/(https:\/\/[^/]+)/)
    const repo = repoMatch ? repoMatch[1] : undefined
    return [buttonPropsForNewRepoSecret(location, { repo, namespace: props.model.namespace, startOrAdd })]
  } else {
    return []
  }
}
