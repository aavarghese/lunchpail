import { taskqueues } from "./Card"
import DrawerContent from "../Drawer/Content"
import DeleteResourceButton from "../DeleteResourceButton"
import { dl as DescriptionList, descriptionGroup } from "../DescriptionGroup"

import Yaml from "../YamlFromObject"
import LinkToNewWizard from "../../navigate/wizard"

import type Props from "@jay/common/events/ApplicationSpecEvent"

/**
 * If we can find a "foo.py", then append it to the repo, so that
 * users can click to see the source directly.
 */
function repoPlusSource(props: Props) {
  const source = props.spec.command.match(/\s(\w+\.py)\s/)
  return props.spec.repo + (source ? "/" + source[1] : "")
}

/** The DescriptionList groups to show in this Detail view */
function detailGroups(props: Props) {
  return Object.entries(props.spec)
    .filter(([, value]) => value)
    .map(([term, value]) =>
      term === "repo"
        ? descriptionGroup(term, repoPlusSource(props))
        : term === "inputs"
        ? taskqueues(props)
        : typeof value !== "function" && typeof value !== "object" && descriptionGroup(term, value),
    )
}

/** Delete this resource */
function deleteAction(props: Props) {
  return (
    <DeleteResourceButton
      kind="applications.codeflare.dev"
      uiKind="applications"
      name={props.metadata.name}
      namespace={props.metadata.namespace}
    />
  )
}

function Edit(props: Props) {
  const qs = [`yaml=${encodeURIComponent(JSON.stringify(props))}`]
  return <LinkToNewWizard startOrAdd="edit" kind="applications" linkText="Edit" qs={qs} />
}

function ApplicationDetail(props: Props) {
  const { inputs } = props.spec
  const otherTabs =
    inputs && inputs.length > 0 && typeof inputs[0].schema === "object"
      ? [
          {
            title: "Test Schema",
            body: <Yaml showLineNumbers={false} obj={JSON.parse(inputs[0].schema.json)} />,
            hasNoPadding: true,
          },
        ]
      : undefined

  return (
    <DrawerContent
      summary={props && <DescriptionList groups={detailGroups(props)} />}
      raw={props}
      otherTabs={otherTabs}
      actions={props && [<Edit {...props} />]}
      rightActions={props && [deleteAction(props)]}
    />
  )
}

export default function MaybeApplicationDetail(props: Props | undefined) {
  return props && <ApplicationDetail {...props} />
}
