import type { ReactNode } from "react"
import { Text } from "@patternfly/react-core"

import names from "../../names"
import Sparkline from "../Sparkline"
import GridLayout from "../GridLayout"
import IconWithLabel from "../IconWithLabel"
import { descriptionGroup } from "../DescriptionGroup"
import { meanCompletionRate, completionRateHistory } from "../CompletionRate"
import { linkToAllApplicationDetails, linkToAllDataSetDetails } from "../../navigate/details"

import type Props from "./Props"
import type { CardHeaderActionsObject } from "@patternfly/react-core"

export function pluralize(text: string, value: number) {
  return `${value} ${text}${value !== 1 ? "s" : ""}`
}

function completionRate(props: Props) {
  return <Sparkline data={completionRateHistory(props.model.events)} />
}

function latestApplications(props: Props) {
  if (props.status) {
    return [props.status.spec.application.name]
  }
  return null
}

function latestDataSets(props: Props) {
  if (props.status) {
    return [props.status.spec.dataset]
  }
  return null
}

function count(props: Props) {
  return !props.status ? 0 : props.status.spec.workers.count
}

/** One row per worker, within row, one cell per inbox or outbox enqueued task */
function enqueued(props: Props) {
  return (
    <div className="codeflare--workqueues">
      {props.model.inbox.map((inbox, i) => (
        <GridLayout key={i} queueNum={i + 1} inbox={inbox} datasetIndex={props.datasetIndex} gridTypeData="plain" />
      ))}
    </div>
  )
}

function numProcessing(props: Props) {
  return (props.model.processing || []).reduce(
    (N: number, processing) => N + Object.values(processing).reduce((M, size) => M + size, 0),
    0,
  )
}

/** "FooBar" -> "Foo Bar" */
export function titleCaseSplit(str: string) {
  return str.split(/(?<=[a-z])(?=[A-Z])|(?<=[A-Z])(?=[A-Z][a-z])/).join(" ")
}

export function statusActions(
  props: Props,
  textComponent?: import("@patternfly/react-core").TextProps["component"],
): CardHeaderActionsObject & { actions: [] | [ReactNode] } {
  const latestStatus = props.status
  const status = latestStatus?.metadata.annotations["codeflare.dev/status"] || "Unknown"
  const message = latestStatus?.metadata.annotations["codeflare.dev/message"]

  return {
    hasNoOffset: true,
    actions: !latestStatus
      ? []
      : [
          <IconWithLabel
            key="Status"
            popoverHeader={titleCaseSplit(status)}
            popoverBody={message}
            status={/Failed/.test(status) ? "Failed" : status}
          >
            <Text component={textComponent}>{titleCaseSplit(status)}</Text>
          </IconWithLabel>,
        ],
  }
}

export function summaryGroups(props: Props) {
  const applications = latestApplications(props)
  const datasets = latestDataSets(props)

  return [
    applications && descriptionGroup(names["applications"], linkToAllApplicationDetails(applications)),
    datasets && descriptionGroup(names["datasets"], linkToAllDataSetDetails(datasets)),
    descriptionGroup("Processing", numProcessing(props)),
    descriptionGroup("Completion Rate", completionRate(props), meanCompletionRate(props.model.events) || "None"),
    descriptionGroup(`Queued Work (${pluralize("worker", count(props))})`, enqueued(props)),
  ]
}
