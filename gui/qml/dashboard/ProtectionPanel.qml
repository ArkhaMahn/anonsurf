import QtQuick
import QtQuick.Layouts

Rectangle {
    id: root

    required property var backend
    property bool safeKill: false
    signal safeKillToggled(bool enabled)
    readonly property bool compact: width < 620
    readonly property bool veryCompact: width < 500
    readonly property color stateColor:
        root.backend.protectionState === "PROTECTED" ? Theme.neonCyan
        : root.backend.protectionState === "DEGRADED" ? Theme.warning
        : root.backend.protectionState === "INACTIVE" ? Theme.textSecondary
        : Theme.danger

    implicitHeight: 104
    color: Theme.bgPanel
    border.width: 1
    border.color: Theme.borderLight

    RowLayout {
        anchors.fill: parent
        anchors.leftMargin: Theme.spacingLarge
        anchors.rightMargin: Theme.spacingLarge
        spacing: Theme.spacingLarge

        Icon {
            Layout.preferredWidth: visible ? 48 : 0
            Layout.preferredHeight: visible ? 48 : 0
            Layout.alignment: Qt.AlignVCenter
            visible: !root.veryCompact
            icon: root.backend.protectionState === "DEGRADED"
                  ? "warning" : "shield"
            color: root.stateColor

            SequentialAnimation on opacity {
                running: root.backend.busy
                loops: Animation.Infinite
                NumberAnimation { to: 0.45; duration: 500 }
                NumberAnimation { to: 1; duration: 500 }
            }
        }

        ColumnLayout {
            Layout.fillWidth: true
            Layout.alignment: Qt.AlignVCenter
            spacing: Theme.spacingSmall

            Text {
                text: "Routing status"
                color: Theme.textSecondary
                font.family: Theme.uiFont
                font.pixelSize: Theme.detailSize
            }

            Text {
                text: root.backend.protectionState
                color: root.stateColor
                font.family: Theme.uiFont
                font.pixelSize: root.compact ? Theme.headingSize + 2 : Theme.titleSize
                font.bold: true
                font.letterSpacing: 2
            }

            // While an action runs the detail line carries its progress: it is
            // the only transient message the interface has, and the panel is
            // where the user is already looking.
            Text {
                Layout.fillWidth: true
                text: root.backend.busy ? root.backend.progress
                                        : root.backend.protectionDetail
                color: root.backend.busy ? Theme.neonCyan : Theme.textSecondary
                font.family: Theme.uiFont
                font.pixelSize: Theme.bodySize
                wrapMode: Text.WordWrap
                visible: !root.compact || root.backend.busy
            }
        }

        ColumnLayout {
            Layout.preferredWidth: root.compact ? 150 : 190
            // A nested layout fills by default, which would silently override
            // the width above and let the button sprawl across the panel.
            Layout.fillWidth: false
            Layout.alignment: Qt.AlignVCenter
            spacing: Theme.spacingSmall

            ActionButton {
                Layout.fillWidth: true
                text: root.backend.anonsurfActive ? "Stop AnonSurf" : "Start AnonSurf"
                iconName: root.backend.anonsurfActive ? "stop" : "start"
                tooltip: root.backend.anonsurfActive
                         ? "Removes the transparent proxy rules and puts the system back on its normal, unprotected routing."
                         : "Installs the transparent proxy rules that push every system connection through Tor."
                primary: true
                enabled: root.backend.statusAvailable && !root.backend.busy
                onClicked: root.backend.anonsurfActive
                           ? root.backend.stop(root.safeKill)
                           : root.backend.start(root.safeKill)
            }

            ToggleOption {
                Layout.alignment: Qt.AlignHCenter
                label: "Use SafeKill"
                tooltip: "Before the routing changes, closes browsers, mail, chat and download applications and wipes their caches with BleachBit, so nothing keeps writing outside Tor."
                accentColor: Theme.warning
                checked: root.safeKill
                enabled: !root.backend.busy
                // A checkbox click overwrites the binding, so it is restored to
                // keep the box following the state the window actually holds.
                onToggled: {
                    root.safeKillToggled(checked);
                    checked = Qt.binding(() => root.safeKill);
                }
            }
        }
    }

    Rectangle {
        anchors.left: parent.left
        anchors.bottom: parent.bottom
        width: root.backend.busy ? parent.width : 0
        height: 2
        color: Theme.neonCyan
        Behavior on width {
            NumberAnimation { duration: Theme.durationMedium }
        }
    }
}
