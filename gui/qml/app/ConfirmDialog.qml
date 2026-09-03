import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Dialog {
    id: root

    property string badge: ""
    property string message: ""
    property string confirmText: "Confirm"
    property color accentColor: Theme.warning
    readonly property int maxWidth: 460

    parent: Overlay.overlay
    anchors.centerIn: parent
    width: parent ? Math.min(root.maxWidth, parent.width - Theme.contentMargin * 2)
                  : root.maxWidth
    modal: true
    focus: true
    padding: Theme.spacingLarge
    standardButtons: Dialog.NoButton
    header: null

    contentItem: ColumnLayout {
        spacing: Theme.spacingMedium

        StatusBadge {
            Layout.alignment: Qt.AlignLeft | Qt.AlignVCenter
            text: root.badge
            accentColor: root.accentColor
        }

        Text {
            Layout.fillWidth: true
            text: root.message
            color: Theme.textSecondary
            font.family: Theme.uiFont
            font.pixelSize: Theme.bodySize
            wrapMode: Text.WordWrap
        }

        RowLayout {
            Layout.fillWidth: true
            Layout.topMargin: Theme.spacingSmall
            spacing: Theme.spacingSmall

            Item { Layout.fillWidth: true }

            ActionButton {
                Layout.preferredWidth: 130
                text: "Cancel"
                iconName: "close"
                onClicked: root.reject()
            }

            ActionButton {
                Layout.preferredWidth: 150
                text: root.confirmText
                iconName: "check"
                primary: true
                onClicked: root.accept()
            }
        }
    }

    background: Rectangle {
        color: Theme.bgPanel
        border.width: Theme.frameBorder
        border.color: root.accentColor
    }
}
