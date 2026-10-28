cmake_minimum_required(VERSION 3.22)
project(anonsurf-gui VERSION 6.0.0 LANGUAGES CXX)

set(ANONSURF_VERSION "${PROJECT_VERSION}" CACHE STRING
    "Version displayed by the AnonSurf GUI"
)
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(Qt6 6.8 REQUIRED COMPONENTS Core DBus Gui Qml Quick QuickControls2 Widgets)
qt_policy(SET QTP0004 NEW)

qt_add_executable(anonsurf-gui
    src/application/main.cpp
    src/application/instance.cpp
    src/application/instance.h
    src/anonsurf/anonsurfactions.cpp
    src/anonsurf/anonsurfbackend.cpp
    src/anonsurf/anonsurfbackend.h
    src/anonsurf/anonsurfstate.cpp
    src/cli/commandclient.cpp
    src/cli/commandclient.h
    src/notifications/notifications.cpp
    src/notifications/notifications.h
    src/systemd/systemdclient.cpp
    src/systemd/systemdclient.h
    src/systemd/systemdjobs.cpp
    src/tray/traycontroller.cpp
    src/tray/traycontroller.h
    src/tray/traymenu.cpp
    src/tray/traymenu.h
)

target_include_directories(anonsurf-gui PRIVATE src)

set_source_files_properties(qml/theme/Theme.qml qml/theme/IconPaths.qml
    PROPERTIES QT_QML_SINGLETON_TYPE TRUE
)

qt_add_qml_module(anonsurf-gui
    URI Parrot.AnonSurf
    VERSION 1.0
    RESOURCE_PREFIX /qt/qml
    QML_FILES
        qml/about/AboutWindow.qml
        qml/about/AboutField.qml
        qml/app/ConfirmDialog.qml
        qml/app/CrtOverlay.qml
        qml/app/ErrorDialog.qml
        qml/app/HeaderBar.qml
        qml/app/Main.qml
        qml/controls/ActionButton.qml
        qml/controls/Icon.qml
        qml/controls/IconButton.qml
        qml/controls/StatusBadge.qml
        qml/controls/StatusItem.qml
        qml/controls/ThemedToolTip.qml
        qml/controls/ToggleOption.qml
        qml/dashboard/ActionPanel.qml
        qml/dashboard/ConnectionOverview.qml
        qml/dashboard/ProtectionPanel.qml
        qml/theme/IconPaths.qml
        qml/theme/Theme.qml
)

set(ANONSURF_ICON "${CMAKE_CURRENT_SOURCE_DIR}/../icons/anonsurf.png")
set_source_files_properties("${ANONSURF_ICON}" PROPERTIES
    QT_RESOURCE_ALIAS icons/anonsurf.png
)
qt_add_resources(anonsurf-gui "anonsurf-assets"
    PREFIX /
    FILES "${ANONSURF_ICON}"
)

target_compile_options(anonsurf-gui PRIVATE
    "$<$<COMPILE_LANG_AND_ID:CXX,GNU,Clang,AppleClang>:-Wall;-Wextra;-Wpedantic>"
)
# A default PIE executable still emits R_X86_64_COPY relocations for external
# data such as Qt's staticMetaObject. Distribution Qt builds (Arch, CachyOS)
# set GNU_PROPERTY_1_NEEDED_INDIRECT_EXTERN_ACCESS on their libraries, and the
# dynamic loader then aborts the program with "copy relocation against
# non-copyable protected symbol". Compiling the executable as fully
# position-independent keeps every external data access GOT-based, and
# -z nocopyreloc makes the linker reject any copy relocation that would sneak
# back in.
set_target_properties(anonsurf-gui PROPERTIES POSITION_INDEPENDENT_CODE ON)
target_compile_options(anonsurf-gui PRIVATE
    "$<$<COMPILE_LANG_AND_ID:CXX,GNU,Clang,AppleClang>:-fPIC>"
)
target_link_options(anonsurf-gui PRIVATE
    "$<$<LINK_LANG_AND_ID:CXX,GNU,Clang,AppleClang>:-Wl,-z,nocopyreloc>"
)
target_compile_definitions(anonsurf-gui PRIVATE
    ANONSURF_VERSION="${ANONSURF_VERSION}"
)
target_link_libraries(anonsurf-gui PRIVATE
    Qt6::Core
    Qt6::DBus
    Qt6::Gui
    Qt6::Qml
    Qt6::Quick
    Qt6::QuickControls2
    Qt6::Widgets
)

install(TARGETS anonsurf-gui RUNTIME DESTINATION bin)
install(FILES org.parrotsec.Anonsurf.desktop DESTINATION share/applications)
install(FILES "${ANONSURF_ICON}"
    DESTINATION share/icons/hicolor/512x512/apps
    RENAME anonsurf.png
)
