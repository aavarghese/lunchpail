import { BrowserTabs } from "@jay/components/S3Browser"
import DrawerContent from "@jay/components/Drawer/Content"
import DeleteResourceButton from "@jay/components/DeleteResourceButton"
import { dl as DescriptionList, descriptionGroup } from "@jay/components/DescriptionGroup"

import { singular } from "../name"
import LinkToNewWizard from "@jay/renderer/navigate/wizard"

import type Props from "./Props"

function detailGroups(props: Props) {
  return Object.entries(props.spec.local)
    .filter(([, value]) => value)
    .map(
      ([term, value]) =>
        typeof value !== "function" &&
        typeof value !== "object" &&
        descriptionGroup(term, value === "COS" ? "S3" : value),
    )
}

/** Delete this resource */
function deleteAction(props: Props) {
  return (
    <DeleteResourceButton
      key="delete"
      kind="dataset"
      singular={singular}
      name={props.metadata.name}
      namespace={props.metadata.namespace}
      context={props.metadata.context}
    />
  )
}

function editAction(props: Props) {
  const qs = [`action=edit&yaml=${encodeURIComponent(JSON.stringify(props))}`]
  return <LinkToNewWizard key="edit" startOrAdd="edit" kind="datasets" linkText="" qs={qs} size="lg" variant="plain" />
}

/** Tabs specific to this kind of data */
function otherTabs(props: Props) {
  const tab = BrowserTabs(props.spec.local)
  return tab ? [tab] : undefined
}

export default function DataSetDetail(props: Props) {
  return (
    <DrawerContent
      summary={<DescriptionList groups={detailGroups(props)} />}
      raw={props}
      rightActions={[editAction(props), deleteAction(props)]}
      otherTabs={otherTabs(props)}
    />
  )
}
