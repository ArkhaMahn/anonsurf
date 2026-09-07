import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import QtQuick.Window

ApplicationWindow {
    id: window

    required property var backend
    property bool safeKill: false
    // Polling costs a process and a D-Bus round trip, so it only runs while
    // the window is on screen.
    readonly property bool live: window.visible
                                 && window.visibility !== Window.Minimized
    // The window opens exactly as tall as its content and cannot grow past it,
    // so no space is ever left blank; making it shorter is allowed and the
    // view scrolls instead.
    readonly property int contentMargins: 2 * (window.width < 600
                                               ? 12 : Theme.contentMargin)
    readonly property int naturalHeight: Math.ceil(contentColumn.implicitHeight
                                                   + headerBar.implicitHeight)
                                         + Theme.spacingMedium
                                         + window.contentMargins
    width: 840
    height: window.naturalHeight
    minimumWidth: 480
    minimumHeight: 240
    maximumHeight: window.naturalHeight
    visible: true
    title: "AnonSurf"
    color: Theme.bgWindow
    font.family: Theme.uiFont

    Component.onCompleted: window.backend.refresh()
    onClosing: aboutWindow.close()

    onLiveChanged: if (window.live) window.backend.refresh()

    Timer {
        interval: 3000
        running: window.live
        repeat: true
        onTriggered: window.backend.refresh()
    }

    Rectangle {
        anchors.fill: parent
        gradient: Gradient {
            GradientStop { position: 0; color: Theme.bgWindow }
            GradientStop { position: 0.62; color: "#071923" }
            GradientStop { position: 1; color: Theme.bgBottom }
        }
    }

    CrtOverlay { anchors.fill: parent }

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: window.width < 600 ? 12 : Theme.contentMargin
        spacing: Theme.spacingMedium

        HeaderBar {
            id: headerBar

            Layout.fillWidth: true
            onAboutRequested: aboutWindow.open()
        }

        ScrollView {
            id: contentScroll

            Layout.fillWidth: true
            Layout.fillHeight: true
            contentWidth: availableWidth
            clip: true

            ScrollBar.horizontal.policy: ScrollBar.AlwaysOff
            ScrollBar.vertical.policy: ScrollBar.AsNeeded

            Column {
                id: contentColumn

                width: contentScroll.availableWidth
                spacing: Theme.spacingMedium

                ProtectionPanel {
                    width: parent.width
                    backend: window.backend
                    safeKill: window.safeKill
                    onSafeKillToggled: enabled => {
                        if (enabled)
                            safeKillDialog.open();
                        else
                            window.safeKill = false;
                    }
                }

                ConnectionOverview {
                    width: parent.width
                    backend: window.backend
                }

                ActionPanel {
                    width: parent.width
                    backend: window.backend
                    safeKill: window.safeKill
                }
            }
        }
    }

    AboutWindow {
        id: aboutWindow
    }

    ConfirmDialog {
        id: safeKillDialog

        badge: "SAFEKILL"
        confirmText: "Enable"
        message: "SafeKill requests termination of selected applications, then "
                 + "cleans selected caches, history and sessions with BleachBit. "
                 + "Unsaved work may be lost and the cleaning cannot be undone."
        onAccepted: window.safeKill = true
    }

    ErrorDialog {
        id: errorDialog

        backend: window.backend
    }

    Connections {
        target: window.backend

        function onChanged() {
            if (window.backend.errorMessage !== "" && !errorDialog.opened)
                errorDialog.open();
        }
    }
}
