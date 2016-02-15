import QtQuick 2.4
import QtQuick.Controls 1.3
import QtQuick.Window 2.2
import QtQuick.Dialogs 1.2
import QtGraphicalEffects 1.0
import QtQml 2.2

import "viewloadercreator.js" as ViewLoaderCreator

ApplicationWindow {
    title: qsTr("Cash Points")
    width: 480
    height: 800
    visible: true

    objectName: "appWindow"
    signal pong(bool ok)
    onPong: {
        // TODO: show user warning about
        // connection to server
        if (ok) {
            console.log("pong :)")
        } else {
            console.warn("pong :(")
        }
    }

    function saveLastGeoPos() {
        var pos = mapView.getMapCenter()
        var zoom = mapView.getMapZoom()
        cashpointModel.saveLastGeoPos(JSON.stringify({
                                          "longitude": pos.longitude,
                                          "latitude": pos.latitude,
                                          "zoom": zoom,
                                      }))
    }

    signal appStateChanged(int state)
    onAppStateChanged: {
        console.log("State Changed:" + state.toString())
        if (state == 2) {
            saveLastGeoPos()
        }
    }

    property date lastExitAttempt: new Date()
    property int backExitThreathold: 500

    property var actions: []

    function handleAction(action, blockSaving) {
        if (!action) {
            return true
        }

        if (action.type === "undo") {
            var lastAction = actions.pop()
            if (!lastAction) {
                return false
            }

            return lastAction.undo(lastAction)
        } else {
            var saveAction = action.do(action)
            if (saveAction && !blockSaving) {
                if (actions.length > 32) {
                    actions.shift()
                }
                actions.push(action)
            }
            return true
        }
    }

    onClosing: {
//        if (Qt.platform.os == "android") {
        {
            console.log("exit")
            var currentTime = new Date()
            if (currentTime.valueOf() - lastExitAttempt.valueOf() < backExitThreathold) {
                saveLastGeoPos()
                close.accepted = true
                return
            }
            lastExitAttempt = currentTime
        }

        var done = handleAction({ "type": "undo" })
        close.accepted = !done
    }

    Keys.onReleased: {
        if (event.key === Qt.Key_Back) {
            event.accepted = true
        }
    }

    Flipable {
        id: flipable
        anchors.fill: parent

        Keys.onEscapePressed: {
            if (flipped) {
                flipped = !flipped
            }
        }

        Keys.onPressed: {
            console.log("here")
            if (event.key === Qt.Key_Tab) {
                if (!flipped) {
                    handleAction({
                                     "type": "flipBack",
                                     "do": function(act) {
                                         flipable.flipped = !flipable.flipped
                                     }
                                 }, true)
                }
            }
        }

        property bool flipped: false
        states: State
                {
                    name: "back"
                    PropertyChanges {
                        target: rotation
                        angle: 180
                    }

                    when: flipable.flipped
                }
        transitions: Transition
                     {
                        NumberAnimation {
                            target: rotation
                            properties: "angle"
                            duration: 1500
                            easing.type: Easing.InOutQuad
                        }
                     }

        transform: [ Rotation
                     {
                        id: rotation
                        origin.x: flipable.width / 2
                        origin.y: flipable.height / 2
                        axis.x: 0; axis.y: 1; axis.z: 0     // set axis.y to 1 to rotate around y-axis
                        angle: 0    // the default angle
                     }
                   ]


        front:
        Item {
            enabled: !parent.flipped
            anchors.fill: parent

            Label {
                id: logo
                anchors.centerIn: parent
                text: "Cash Points"
                font.pixelSize: 96
                color: "blue"
                fontSizeMode: Text.HorizontalFit
                property bool initialized: false

                MouseArea {
                    anchors.fill: parent
                    onClicked:
                        if (logo.state == "")
                        {
                            logo.state = "animationStart"
                            flipable.flipped = !flipable.flipped
                        } else {
                            console.warn("test")
                        }
                }

                states: [
                    State {
                        name: "animationStart"
                        PropertyChanges {
                            target: logo
                            color: "lightgray"
                        }
                        onCompleted: logo.state = "animationEnd"
                    },
                    State {
                        name: "animationEnd"
                        PropertyChanges {
                            target: logo
                            color: "black"
                        }
                        onCompleted: logo.state = "animationStart"
                    } ]

                transitions: [
                    Transition {
                        from: ""
                        to: "animationStart"

                        ColorAnimation {
                            target: logo
                            duration: 2000
                        }
                    },
                    Transition {
                        from: "animationStart"
                        to: "animationEnd"

                        ColorAnimation {
                            target: logo
                            duration: 2000
                        }
                    },
                    Transition {
                        from: "animationEnd"
                        to: "animationStart"

                        ColorAnimation {
                            target: logo
                            duration: 2000
                        }
                    },
                    Transition {
                        from: "*"
                        to: ""

                        ColorAnimation {
                            target: logo
                            duration: 2000
                        }
                    }
                ]
            }

            BusyIndicator {
                id: progress
                //anchors.centerIn: parent
                anchors.top: logo.bottom
                anchors.horizontalCenter: logo.horizontalCenter
//                value: 0.8
            }

//            ProgressBar
//            {
//                id: progress
//                //anchors.centerIn: parent
//                anchors.top: logo.bottom
//                anchors.horizontalCenter: logo.horizontalCenter
//                value: 0.8
//            }
        }
        back:
        MapView {
            id: mapView
            enabled: parent.flipped
            anchors.fill: parent
            showControls: leftMenu.state == "hidden"
            active: !leftMenu.visible

            onAction: {
                handleAction(action)
            }

            LeftMenu {
                id: leftMenu
                x: 0
                z: mapView.z + 10
                anchors.top: parent.top
                anchors.bottom: parent.bottom
                width: parent.width * 0.6

                onItemClicked: {
                    if (itemName && itemName.length > 0) {
                        handleAction({
                                         "type": "openView",
                                         "do": function(act) {
                                             ViewLoaderCreator.createViewLoader(function(loader) {
                                                 loader.setView(itemName)
                                             })
                                             flipable.flipped = !flipable.flipped
                                             return true
                                         },
                                         "undo": function(act) {
                                             flipable.flipped = !flipable.flipped
                                             return true
                                         }
                                     })
                    }
                }
            }

            RectangularGlow {
                id: leftMenuGlow
                visible: leftMenu.visible
                anchors.fill: leftMenu
                z: leftMenu.z - 1
                glowRadius: leftMenu.height / 10
                spread: 0.1
                color: "#0000000FF"
                cornerRadius: glowRadius
                opacity: (leftMenu.x + leftMenu.width) / (mapView.width * 0.6)
            }

            RectangularGlow {
                visible: leftMenu.visible
                anchors.fill: mapView
                z: leftMenu.z - 1
                color: "#0000000FF"
                glowRadius: 100000
                opacity: leftMenuGlow.opacity
            }

            onClicked: {
                if (leftMenu.state === "") {
                    handleAction({
                                     "type": "hideMenu",
                                     "do": function(act) {
                                         leftMenu.state = "hidden"
                                         return false
                                     },
                                     "undo": function(act) {
                                         return true
                                     }
                                 })
                }
            }

            onMenuClicked: {
                handleAction({
                                 "type": "showMenu",
                                 "do": function(act) {
                                     leftMenu.state = ""
                                     return true
                                 },
                                 "undo": function(act) {
                                     leftMenu.state = "hidden"
                                     return true
                                 }
                             })
            }
        }
    }
}
