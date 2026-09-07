import QtQuick
import QtQuick.Layouts

Item {
    id: root

    required property var backend
    property bool safeKill: false
    // The row never wraps: below the width that fits five labels the buttons
    // drop to icons, which keeps the panel one line tall at any size instead
    // of stacking and dragging the whole window taller.
    readonly property bool iconOnly: width < 700

    // No frame here: five buttons in a row already read as one group, and a
    // container around them would only add a box inside a box.
    implicitHeight: content.implicitHeight

    ColumnLayout {
        id: content

        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        spacing: Theme.spacingSmall

        GridLayout {
            Layout.fillWidth: true
            columns: 5
            columnSpacing: Theme.spacingSmall
            rowSpacing: Theme.spacingSmall

            ActionButton {
                Layout.fillWidth: true
                text: "New Identity"
                iconName: "identity"
                iconOnly: root.iconOnly
                tooltip: "Asks Tor for a fresh circuit, so the exit address websites see changes."
                enabled: root.backend.statusAvailable && root.backend.anonsurfActive
                         && root.backend.torActive && !root.backend.busy
                onClicked: root.backend.changeIdentity()
            }

            ActionButton {
                Layout.fillWidth: true
                text: "Verify IP"
                iconName: "globe"
                iconOnly: root.iconOnly
                tooltip: "Reads the current public address and reports whether it belongs to a known Tor exit node."
                enabled: root.backend.statusAvailable && !root.backend.busy
                onClicked: root.backend.checkIP()
            }

            ActionButton {
                Layout.fillWidth: true
                text: "Restart"
                iconName: "restart"
                iconOnly: root.iconOnly
                tooltip: "Tears down and rebuilds the routing rules, keeping every setting as it is."
                enabled: root.backend.statusAvailable && root.backend.anonsurfActive
                         && !root.backend.busy
                onClicked: root.backend.restart(root.safeKill)
            }

            ActionButton {
                Layout.fillWidth: true
                text: "Tor Monitor"
                iconName: "monitor"
                iconOnly: root.iconOnly
                tooltip: "Opens nyx, the Tor connection monitor, in a Konsole window."
                enabled: root.backend.statusAvailable && root.backend.anonsurfActive
                         && root.backend.torActive && !root.backend.busy
                onClicked: root.backend.openMonitor()
            }

            ActionButton {
                Layout.fillWidth: true
                text: "Refresh"
                iconName: "refresh"
                iconOnly: root.iconOnly
                tooltip: "Reads the AnonSurf status right away instead of waiting for the periodic update."
                enabled: !root.backend.busy
                onClicked: root.backend.refresh()
            }
        }

    }
}
