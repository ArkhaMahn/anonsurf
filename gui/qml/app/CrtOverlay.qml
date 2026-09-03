pragma ComponentBehavior: Bound
import QtQuick

Item {
    id: root

    Repeater {
        model: Math.ceil(root.height / Theme.scanlineStep)

        Rectangle {
            required property int index

            y: index * Theme.scanlineStep
            width: root.width
            height: 1
            color: Theme.scanline
            opacity: 0.38
        }
    }

    Rectangle {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        height: parent.height * 0.16
        gradient: Gradient {
            GradientStop { position: 0; color: Theme.vignette }
            GradientStop { position: 1; color: "transparent" }
        }
    }

    Rectangle {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.bottom: parent.bottom
        height: parent.height * 0.2
        gradient: Gradient {
            GradientStop { position: 0; color: "transparent" }
            GradientStop { position: 1; color: Theme.vignette }
        }
    }
}
