import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Dialog {
    id: root

    required property var backend
    readonly property int maxWidth: 520

    parent: Overlay.overlay
    anchors.centerIn: parent
    width: parent ? Math.min(root.maxWidth, parent.width - Theme.contentMargin * 2)
                  : root.maxWidth
    height: parent ? Math.min(260, parent.height - Theme.contentMargin * 2) : 260
    modal: true
    focus: true
    padding: Theme.spacingLarge
    standardButtons: Dialog.NoButton
    title: "Action failed"
    header: null
    onClosed: root.backend.dismissError()

    contentItem: ColumnLayout {
        spacing: Theme.spacingMedium

        StatusBadge {
            Layout.alignment: Qt.AlignLeft | Qt.AlignVCenter
            text: "SYSTEM ERROR"
            accentColor: Theme.danger
        }

        ScrollView {
            id: errorScroll

            Layout.fillWidth: true
            Layout.fillHeight: true
            contentWidth: availableWidth
            clip: true
            ScrollBar.horizontal.policy: ScrollBar.AlwaysOff

            Text {
                width: errorScroll.availableWidth
                text: root.backend.errorMessage
                color: Theme.textSecondary
                font.family: Theme.uiFont
                font.pixelSize: Theme.bodySize
                wrapMode: Text.WordWrap
            }
        }

        ActionButton {
            Layout.alignment: Qt.AlignRight
            Layout.preferredWidth: 130
            text: "Dismiss"
            iconName: "close"
            onClicked: root.close()
        }
    }

    background: Rectangle {
        color: Theme.bgPanel
        border.width: Theme.frameBorder
        border.color: Theme.danger
    }
}
