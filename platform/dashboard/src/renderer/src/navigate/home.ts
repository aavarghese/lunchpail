import { hash } from "./kind"
import type { LocationProps } from "../router/withLocation"

function returnTo(props: LocationProps, hash = props.location.hash) {
  const returnTo = props.searchParams.get("returnTo")
  const to = returnTo ? decodeURIComponent(returnTo).replace(/#\w+/, "") : props.location.pathname
  props.navigate(to + hash)
}

export default function navigateToHome(props: LocationProps) {
  returnTo(props)
}

export function navigateToWorkerPools(props: LocationProps) {
  returnTo(props, hash("workerpools"))
}