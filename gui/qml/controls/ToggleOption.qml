import QtQuick
import QtQuick.Controls

CheckBox {
    id: root

    property alias label: textLabel.text
    property string tooltip: ""
    property color accentColor: Theme.accentBright

    spacing: Theme.spacingSmall
    hoverEnabled: true
    Accessible.name: label

    indicator: Rectangle {
        x: 0
        y: parent.height / 2 - height / 2
        width: 20
        height: 20
        color: root.checked ? root.accentColor : Theme.bgPanelAlt
        border.width: 1
        border.color: root.checked ? root.accentColor : Theme.textSecondary

        Icon {
            anchors.centerIn: parent
            width: 12
            height: 12
            icon: "check"
            color: Theme.bgWindow
            visible: root.checked
        }
    }

    contentItem: Text {
        id: textLabel

        leftPadding: 28
        verticalAlignment: Text.AlignVCenter
        color: !root.enabled ? Theme.borderLight
                             : root.checked ? root.accentColor : Theme.textSecondary
        font.family: Theme.uiFont
        font.pixelSize: Theme.detailSize
    }

    ThemedToolTip {
        parent: root
        x: 0
        y: root.height + Theme.spacingTiny
        visible: root.tooltip !== "" && root.hovered
        text: root.tooltip
    }
}
