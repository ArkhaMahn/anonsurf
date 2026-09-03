import QtQuick
import QtQuick.Layouts

Item {
    id: root

    signal aboutRequested()
    implicitHeight: 50

    RowLayout {
        anchors.fill: parent
        spacing: Theme.spacingSmall

        Item {
            Layout.preferredWidth: 44
            Layout.preferredHeight: 44
            Layout.alignment: Qt.AlignVCenter

            Image {
                anchors.centerIn: parent
                width: 42
                height: 42
                source: "qrc:/icons/anonsurf.png"
                fillMode: Image.PreserveAspectFit
            }
        }

        ColumnLayout {
            Layout.fillWidth: true
            Layout.alignment: Qt.AlignVCenter
            spacing: Theme.spacingTiny

            Text {
                text: "ANONSURF"
                color: Theme.textPrimary
                font.family: Theme.uiFont
                font.pixelSize: Theme.headingSize
                font.bold: true
                font.letterSpacing: 1
            }

            Text {
                Layout.fillWidth: true
                text: "Route system traffic through Tor"
                color: Theme.textSecondary
                font.family: Theme.uiFont
                font.pixelSize: Theme.detailSize
                elide: Text.ElideRight
            }
        }

        IconButton {
            Layout.alignment: Qt.AlignVCenter
            iconName: "info"
            tooltip: "About AnonSurf"
            onClicked: root.aboutRequested()
        }
    }
}
