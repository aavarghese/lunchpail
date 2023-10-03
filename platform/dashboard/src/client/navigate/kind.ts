import type { Kind } from "../names"
import type { LocationProps } from "../router/withLocation"

const defaultKind: Kind = "datasets"

export function hash(kind: Kind) {
  return "#" + kind
}

/**
 * Avoid an extra # in the URI if we are navigating to the
 * defaultKind.
 */
export function hashIfNeeded(kind: Kind) {
  return kind === defaultKind ? "#" : "#" + kind
}

export function currentKind(props: Pick<LocationProps, "location">): Kind {
  return (props.location.hash.slice(1) as Kind) || defaultKind
}

export default function isShowingKind(
  kind: "applications" | "datasets" | "workerpools",
  props: Pick<LocationProps, "location">,
) {
  return kind === currentKind(props)
}
