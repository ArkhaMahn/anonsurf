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
            tooltip: "Tor service state. ACTIVE alone does not confirm that a request uses Tor."
            value: !root.backend.statusAvailable ? "UNKNOWN"
                   : root.backend.torActive ? "ACTIVE" : "INACTIVE"
            detail: !root.backend.statusAvailable ? "Unavailable"
                    : root.backend.anonsurfActive ? "Managed by AnonSurf" : "System service"
            accentColor: root.backend.statusAvailable && root.backend.torActive
                         ? Theme.neonCyan : Theme.textPrimary
        }

        StatusItem {
            Layout.fillWidth: true
            label: "DNS setup"
            tooltip: "DNS configuration inferred from the resolver and service state, not a live DNS routing check."
            value: root.backend.dnsMode
            detail: root.backend.dnsServers
            accentColor: !root.backend.statusAvailable || !root.backend.anonsurfActive
                         ? Theme.textPrimary
                         : root.backend.dnsMode.indexOf("Tor") === 0
                           ? Theme.neonCyan : Theme.warning
        }

        StatusItem {
            Layout.fillWidth: true
            label: "Public IP"
            tooltip: "IP observed by the Tor Project for this request. A Tor exit result does not verify every application's traffic."
            value: root.backend.publicIP
            detail: "Tor check: " + root.backend.torCheck
            accentColor: root.backend.torCheck === "TOR" ? Theme.neonCyan
                         : root.backend.torCheck === "CLEARNET" ? Theme.danger
                         : Theme.textPrimary
        }

        Item {
            Layout.fillWidth: true
            Layout.alignment: Qt.AlignTop
            implicitHeight: startupContent.implicitHeight

            ColumnLayout {
                id: startupContent

                anchors.left: parent.left
                anchors.right: parent.right
                anchors.top: parent.top
                spacing: Theme.spacingTiny

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
                    tooltip: "Starts AnonSurf during boot. Traffic sent before routing is activated is not covered."
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
