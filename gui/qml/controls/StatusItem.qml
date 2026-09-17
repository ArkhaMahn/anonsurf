import QtQuick
import QtQuick.Layouts

Item {
    id: root

    property string label: ""
    property string value: ""
    property string detail: ""
    property string tooltip: ""
    property color accentColor: Theme.neonCyan

    implicitHeight: content.implicitHeight

    HoverHandler {
        id: hover

        enabled: root.tooltip !== ""
    }

    ThemedToolTip {
        parent: root
        x: 0
        y: root.height + Theme.spacingTiny
        visible: hover.hovered
        text: root.tooltip
    }

    ColumnLayout {
        id: content

        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        spacing: Theme.spacingTiny

        Text {
            text: root.label
            color: Theme.textSecondary
            font.family: Theme.uiFont
            font.pixelSize: Theme.badgeSize
            font.weight: Font.Medium
        }

        Text {
            Layout.fillWidth: true
            text: root.value
            color: root.accentColor
            font.family: Theme.uiFont
            font.pixelSize: Theme.bodySize
            font.bold: true
            elide: Text.ElideRight
        }

        Text {
            Layout.fillWidth: true
            text: root.detail
            color: Theme.textSecondary
            font.family: Theme.uiFont
            font.pixelSize: Theme.badgeSize
            elide: Text.ElideMiddle
        }
    }
}
