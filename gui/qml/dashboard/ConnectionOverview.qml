import QtQuick
import QtQuick.Layouts

Rectangle {
    id: root

    required property var backend
    // Four columns hold down to the width where each one would stop fitting a
    // readable value; below that two, never a single stacked column, which
    // would make the window grow much taller than it needs to be.
    readonly property int columns: width < 560 ? 2 : 4

    implicitHeight: content.implicitHeight + Theme.spacingMedium * 2
    color: Theme.bgPanel
    border.width: 1
    border.color: Theme.borderLight

    GridLayout {
        id: content

        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        anchors.margins: Theme.spacingMedium
        columns: root.columns
        columnSpacing: Theme.spacingLarge
        rowSpacing: Theme.spacingSmall

        StatusItem {
            Layout.fillWidth: true
            label: "Tor service"
            tooltip: "State of the tor daemon that actually carries the traffic. AnonSurf takes it over while routing is active."
            value: !root.backend.statusAvailable ? "UNKNOWN"
                   : root.backend.torActive ? "ACTIVE" : "INACTIVE"
            detail: !root.backend.statusAvailable ? "Unavailable"
                    : root.backend.anonsurfActive ? "Managed by AnonSurf" : "System service"
            accentColor: root.backend.statusAvailable && root.backend.torActive
                         ? Theme.neonCyan : Theme.textSecondary
        }

        StatusItem {
            Layout.fillWidth: true
            label: "DNS route"
            tooltip: "Where name lookups are sent. Only a Tor route keeps them from revealing to your provider which sites you visit."
            value: root.backend.dnsMode
            detail: root.backend.dnsServers
            accentColor: !root.backend.statusAvailable || !root.backend.anonsurfActive
                         ? Theme.textSecondary
                         : root.backend.dnsMode.indexOf("Tor") === 0
                           ? Theme.neonCyan : Theme.warning
        }

        StatusItem {
            Layout.fillWidth: true
            label: "Public IP"
            tooltip: "Address the sites you visit see. The Tor check tells whether it belongs to a known exit node."
            value: root.backend.publicIP
            detail: "Tor check: " + root.backend.torCheck
            accentColor: root.backend.torCheck === "TOR" ? Theme.neonCyan
                         : root.backend.torCheck === "CLEARNET" ? Theme.danger
                         : Theme.textSecondary
        }

        Item {
            Layout.fillWidth: true
            implicitHeight: 54

            ColumnLayout {
                anchors.fill: parent
                spacing: 2

                Text {
                    text: "System startup"
                    color: Theme.textSecondary
                    font.family: Theme.uiFont
                    font.pixelSize: Theme.badgeSize
                    font.weight: Font.Medium
                }

                ToggleOption {
                    Layout.fillWidth: true
                    label: root.backend.bootEnabled ? "Enabled" : "Disabled"
                    tooltip: "Enables the AnonSurf service at boot, so routing is already in place when the session starts."
                    checked: root.backend.bootEnabled
                    enabled: !root.backend.busy
                    // The box follows the service, not the click: a refused
                    // authorisation leaves the real state on screen.
                    onToggled: {
                        if (checked !== root.backend.bootEnabled)
                            root.backend.setBootEnabled(checked);
                        checked = Qt.binding(() => root.backend.bootEnabled);
                    }
                }
            }
        }
    }
}
