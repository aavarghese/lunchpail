// @ts-check
import { ElectronApplication, Page, expect, test } from "@playwright/test"
import launchElectron from "./launch-electron"

test.describe.serial("workers tests running sequentially", () => {
  let electronApp: ElectronApplication
  let page: Page
  let demoModeStatus: boolean
  const expectedApp = "worm"
  const expectedTaskQueue = "purple"

  test("Task Queue link opens associated drawer", async () => {
    // Launch Electron app.
    electronApp = await launchElectron()

    // Get the first page that the app opens, wait if necessary.
    page = await electronApp.firstWindow()

    // Check if we are in demo mode (should be true by default)
    demoModeStatus = await page.getByLabel("Demo").isChecked()
    console.log(`Demo mode on?: ${demoModeStatus}`)

    // get Applications tab element from the sidebar and click to activate Application gallery
    await page.getByRole("link", { name: "Code" }).click()

    // click on the task queue link of one of the cards
    await page.getByRole("link", { name: expectedTaskQueue }).click()

    // check that the drawer for that task queue view opened
    const id = "taskqueues." + expectedTaskQueue
    const drawer = await page.locator(`[data-ouia-component-id="${id}"]`)
    await expect(drawer).toBeVisible()

    // verify that the drawer that opened matches the task queue that was clicked
    const drawerTitle = await drawer.locator(`[data-ouia-component-type="PF5/Title"]`)
    await expect(drawerTitle).toContainText(expectedTaskQueue)
  })

  test("'Assign Workers' button opens 'Create Compute Pool' modal", async () => {
    // click on 'assign workers' button
    await page.getByRole("link", { name: "Assign Workers" }).click()

    // check that modal opened
    const modal = await page.locator(`[data-ouia-component-type="PF5/ModalContent"]`)
    await expect(modal).toBeVisible()
  })

  test("'Create Compute Pool' modal is autopopulated", async () => {
    // check that 'Application Code' drop down matches expectedApp
    // const appCodeMenuToggle = await page.locator(`[.pf-v5-c-menu-toggle][0]`)
    await expect(page.getByRole("button", { name: expectedApp })).toBeVisible()

    // check that 'Task Queue' drop down matches expectedTaskQueue
    // const taskQueueMenuToggle = await page.locator(`[.pf-v5-c-menu-toggle][1]`)
    await expect(page.getByRole("button", { name: expectedTaskQueue })).toBeVisible()
  })

  test("Clicking 'Next' and 'Register Compute Pool' in modal", async () => {
    // click 'Next' and verify that we moved on to 'Review' window
    await page.getByRole("button", { name: "Next" }).click()
    const modalPage = await page.locator(`.pf-v5-c-wizard__toggle`)
    await expect(modalPage).toContainText("Review")

    // click 'Register Compute Pool'
    await page.getByRole("button", { name: "Register Compute Pool" }).click()

    // // Check that the Drawer updated with new worker information
    // const drawer = await page.locator(`[data-ouia-component-id="PF5/DrawerPanelContent"]`)
    // workerName = await drawer.locator(`data-ouia-component-type="PF5/Title"`).innerText()
    // const workerDrawer = await page.locator(`[data-ouia-component-id="workerpools.${workerName}"]`)
    // await expect(workerDrawer).toBeDefined()

    // // then extract the worker name
    // await expect(workerDrawer.locator(`[data-ouia-component-type="PF5/Breadcrumb"]`)).toContainText(["Resources", "Compute"])
  })
})
