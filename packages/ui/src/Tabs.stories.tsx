import type { Meta, StoryObj } from '@storybook/react'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './Tabs'

const meta = {
  title: 'Components/Tabs',
  component: Tabs,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Tabs>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Tabs defaultValue="account">
      <TabsList>
        <TabsTrigger value="account">Account</TabsTrigger>
        <TabsTrigger value="password">Password</TabsTrigger>
        <TabsTrigger value="notifications">Notifications</TabsTrigger>
      </TabsList>
      <TabsContent value="account">
        <div className="p-4 rounded-lg border border-border">
          <h3 className="text-sm font-semibold mb-1">Account Settings</h3>
          <p className="text-sm text-muted-foreground">
            Manage your account details and preferences here.
          </p>
        </div>
      </TabsContent>
      <TabsContent value="password">
        <div className="p-4 rounded-lg border border-border">
          <h3 className="text-sm font-semibold mb-1">Change Password</h3>
          <p className="text-sm text-muted-foreground">
            Update your password to keep your account secure.
          </p>
        </div>
      </TabsContent>
      <TabsContent value="notifications">
        <div className="p-4 rounded-lg border border-border">
          <h3 className="text-sm font-semibold mb-1">
            Notification Preferences
          </h3>
          <p className="text-sm text-muted-foreground">
            Configure how and when you receive notifications.
          </p>
        </div>
      </TabsContent>
    </Tabs>
  ),
}

export const Vertical: Story = {
  render: () => (
    <Tabs defaultValue="account" orientation="vertical">
      <div className="flex gap-4">
        <TabsList>
          <TabsTrigger value="account">Account</TabsTrigger>
          <TabsTrigger value="password">Password</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
        </TabsList>
        <div className="flex-1">
          <TabsContent value="account">
            <div className="p-4 rounded-lg border border-border">
              <h3 className="text-sm font-semibold mb-1">Account Settings</h3>
              <p className="text-sm text-muted-foreground">
                Manage your account details and preferences here.
              </p>
            </div>
          </TabsContent>
          <TabsContent value="password">
            <div className="p-4 rounded-lg border border-border">
              <h3 className="text-sm font-semibold mb-1">Change Password</h3>
              <p className="text-sm text-muted-foreground">
                Update your password to keep your account secure.
              </p>
            </div>
          </TabsContent>
          <TabsContent value="notifications">
            <div className="p-4 rounded-lg border border-border">
              <h3 className="text-sm font-semibold mb-1">
                Notification Preferences
              </h3>
              <p className="text-sm text-muted-foreground">
                Configure how and when you receive notifications.
              </p>
            </div>
          </TabsContent>
        </div>
      </div>
    </Tabs>
  ),
}

export const Disabled: Story = {
  render: () => (
    <Tabs defaultValue="account">
      <TabsList>
        <TabsTrigger value="account">Account</TabsTrigger>
        <TabsTrigger value="password" disabled>
          Password (disabled)
        </TabsTrigger>
        <TabsTrigger value="notifications">Notifications</TabsTrigger>
      </TabsList>
      <TabsContent value="account">
        <div className="p-4 rounded-lg border border-border">
          <p className="text-sm text-muted-foreground">Account tab content.</p>
        </div>
      </TabsContent>
      <TabsContent value="password">
        <div className="p-4 rounded-lg border border-border">
          <p className="text-sm text-muted-foreground">Password tab content.</p>
        </div>
      </TabsContent>
      <TabsContent value="notifications">
        <div className="p-4 rounded-lg border border-border">
          <p className="text-sm text-muted-foreground">
            Notifications tab content.
          </p>
        </div>
      </TabsContent>
    </Tabs>
  ),
}
