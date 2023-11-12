import { useLocation, useSearchParams } from "react-router-dom"

import {
  Alert,
  AlertActionLink,
  Badge,
  DrawerPanelBody,
  Stack,
  Tabs,
  Tab,
  TabAction,
  TabTitleIcon,
  TabTitleText,
} from "@patternfly/react-core"

import taskqueueProps from "../taskqueueProps"
import WorkerPoolIcon from "../../../workerpools/components/Icon"
import { summaryTabContent as queueTabContent } from "../../../taskqueues/components/Detail"
import { name as workerpoolName, singular as workerpoolSingular } from "../../../workerpools/name"
import { correctiveLinks, summaryTabContent as computeTabContent } from "../../../workerpools/components/Detail"

import type Props from "../Props"

/** Tab that shows Compute */
export default function computeTab(props: Props) {
  const location = useLocation()
  const [searchParams] = useSearchParams()

  const queueProps = taskqueueProps(props)
  const models = props.latestWorkerPoolModels.filter((_) => _.application === props.application.metadata.name)
  if (!queueProps) {
    return []
  }

  const computeBody =
    models.length === 0 ? (
      <></>
    ) : (
      <Tabs mountOnEnter defaultActiveKey={models[0].label}>
        {models.map((model) => {
          const workerpoolProps: import("../../../workerpools/components/Props").default = {
            model,
            taskqueueIndex: props.taskqueueIndex,
            status: props.workerpools.find((_) => models[0].label === _.metadata.name),
          }

          const corrections = correctiveLinks({ location, searchParams }, workerpoolProps)
          const tabBody = (
            <Stack hasGutter>
              {corrections.length > 0 && (
                <Alert
                  isInline
                  variant="danger"
                  title={`Unhealthy ${workerpoolSingular}`}
                  actionLinks={corrections.map((_) => (
                    <AlertActionLink {..._} />
                  ))}
                >
                  This {workerpoolSingular} is unhealthy. Consider taking the suggested corrective action
                  {corrections.length === 1 ? "" : "s"}.
                </Alert>
              )}
              {computeTabContent(workerpoolProps, true)}
            </Stack>
          )

          return (
            <Tab
              key={model.label}
              title={
                <>
                  <TabTitleIcon>
                    <WorkerPoolIcon />
                  </TabTitleIcon>
                  <TabTitleText>{model.label.replace(queueProps.name + "-pool-", "")}</TabTitleText>
                </>
              }
              eventKey={model.label}
            >
              <DrawerPanelBody>{tabBody}</DrawerPanelBody>
            </Tab>
          )
        })}
      </Tabs>
    )

  const body = (
    <Stack>
      <DrawerPanelBody>{queueTabContent(queueProps, true)}</DrawerPanelBody>
      {computeBody}
    </Stack>
  )

  return [
    {
      title: workerpoolName,
      body,
      hasNoPadding: true,
      actions: (
        <TabAction>
          <Badge isRead={models.length === 0}>{pluralize("worker", models.length)}</Badge>
        </TabAction>
      ),
    },
  ]
}

function pluralize(text: string, value: number) {
  return `${value} ${text}${value !== 1 ? "s" : ""}`
}
