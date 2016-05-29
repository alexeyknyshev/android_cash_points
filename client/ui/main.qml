import QtQuick 2.4
import QtQuick.Controls 1.3
import QtQuick.Window 2.2
import QtQuick.Dialogs 1.2
import QtQml 2.2

import "viewloadercreator.js" as ViewLoaderCreator

ApplicationWindow {
    id: appWindow
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

    property bool banksReceived: false
    property bool townsReceived: false

    property var banksProgress: {
        "done": 0,
        "total": 0,
    }

    property var townsProgress: {
        "done": 0,
        "total": 0,
    }

    property var mainView

    signal serverDataReceived(bool ok, string data)
    onServerDataReceived: {
        var text = ""
        if (ok) {
            console.log("received " + data + " server data")
            if (data === "towns") {
                townsReceived = true
            } else if (data === "banks") {
                banksReceived = true
            }

            if (banksReceived && townsReceived) {
                handleAction({
                                 "type": "openView",
                                 "mode": "replace",
                                 "path": "MainView.qml",
                                 "actionCallback": function(act) {
                                     handleAction(act, {})
                                 },
                                 "initCallback": function(view) {
                                     mainView = view
                                 }
                             })
            }

            /*if (data === "towns") {
                townsReceived = true
                if (!banksReceived) {
                    text = qsTr("Загружаем банки")
                } else {
                    text = qsTr("Почти готово")
                }
            } else if (data === "banks") {
                banksReceived = true
                if (!townsReceived) {
                    text = qsTr("Загружаем города")
                } else {
                    text = qsTr("Почти готово")
                }
            } else {
                text = qsTr("Почти готово")
            }*/

            //progress.setTextInfo(text)

//            if (banksReceived) {
//                    flipable.flipped = true
//            }
        } else {
            text = qsTr("Ошибка подключения к серверу")
            if (data) {
                text += ":\n" + data
            }
            progress.setTextError(text)
            settingButton.opacity = 1.0
            console.log("cannot receive server data")
        }
    }

    signal banksUpdateProgress(int done, int total)
    onBanksUpdateProgress: {
        banksProgress.done = done
        banksProgress.total = total
        updateProgress()
        console.log("banks update progress: " + done.toString() + "/" + total.toString())
    }

    signal townsUpdateProgress(int done, int total)
    onTownsUpdateProgress: {
        townsProgress.done = done
        townsProgress.total = total
        updateProgress()
        console.log("towns update progress: " + done.toString() + "/" + total.toString())
    }

    function updateProgress() {
        var text = ""
        if (banksProgress.done > 0) {
            var bpercent = banksProgress.done / banksProgress.total
            text += qsTr("Банки") + ": " + Math.round(bpercent.toString() * 100) + "%\n"
        }
        if (townsProgress.done > 0) {
            var tpercent = townsProgress.done / townsProgress.total
            text += qsTr("Города") + ": " + Math.round(tpercent.toString() * 100) + "%\n"
        }
        if (text.length > 0) {
            text = qsTr("Загрузка\n") + text
        } else {
            textпить = qsTr("Подготовка")
        }
        progress.setTextInfo(text)
    }

    function saveLastGeoPos() {
        if (mainView) {
            var pos = mainView.getMapCenter()
            var zoom = mainView.getMapZoom()
            cashpointModel.saveLastGeoPos(JSON.stringify({
                                              "longitude": pos.longitude,
                                              "latitude": pos.latitude,
                                              "zoom": zoom,
                                          }))
        }
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

    property var actionCallback

    property var actions: []

    function handleAction(action, opt) {
        if (!action) {
            return true
        }

        if (!opt) {
            opt = {}
        }

        /*if (actionCallback) {
            var ok = actionCallback(action)
            if (!ok) {
                return true
            }
        }*/

        if (action.type === "openView") {
            if (!action.mode) {
                action.mode = "push"
            }

            action.do = function(act) {
                if (action.path) {
                    ViewLoaderCreator.newLoader(function(object) {
                        action.object = object
                        object.setView(action.path, action)

                        var opt = { "item": object }
                        if (action.mode === "replace") {
                            opt.replace = true
                        }
                        mainStack.push(opt)
                    }, action)
                }
                return true
            }
            action.undo = function(act) {
                if (action.onClose) {
                    action.onClose()
                }

                if (action.object) {
                    if (mainStack.depth > 1) {
                        var object = mainStack.pop()
                        if (object) {
                            object.setView("")
                        }
                    }
                }
                return true
            }
        }

        if (action.type === "undo") {
            var lastAction = actions.pop()
            if (!lastAction) {
                return false
            }

            return lastAction.undo(lastAction)
        } else {
            var saveAction = action.do(action)
            if (saveAction && !opt.blockSaving) {
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

    StackView {
        id: mainStack
        anchors.fill: parent

        delegate: StackViewDelegate {
            function transitionFinished(properties)
            {
                properties.exitItem.opacity = 1
            }

            pushTransition: StackViewTransition {
                PropertyAnimation {
                    duration: 800
                    target: enterItem
                    property: "opacity"
                    easing.type: Easing.InQuad
                    from: 0
                    to: 1
                }
                /*PropertyAnimation {
                    duration: 800
                    target: exitItem
                    property: "opacity"
                    easing.type: Easing.OutQuad
                    from: 1
                    to: 0
                }*/
            }

            popTransition: StackViewTransition {
                PropertyAnimation {
                    duration: 800
                    target: exitItem
                    property: "opacity"
                    easing.type: Easing.OutQuad
                    from: 1
                    to: 0
                }
            }
        }

        /*Keys.onEscapePressed: {
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
        }*/

        /*property bool flipped: false
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
                   ]*/


        //front:
        initialItem: Item {
            //enabled: !parent.flipped
            anchors.fill: parent

            Image {
                id: logo
                x: parent.width * 0.2
                y: parent.height * 0.25
                width: parent.width * 0.6
                fillMode: Image.PreserveAspectFit
                source: "qrc:/app_ico.png"
            }

/*
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
*/
            Label {
                id: progress
                //anchors.centerIn: parent
                anchors.top: logo.bottom
                anchors.topMargin: Math.min(parent.height, parent.width) * 0.02
                anchors.horizontalCenter: logo.horizontalCenter
                text: qsTr("Загружаем банки и города")
                font.bold: true
                font.pixelSize: 24
                color: "gray"

                Behavior on opacity {
                    NumberAnimation {
                        duration: 1000
                    }
                }

                function setTextInfo(newText) {
                    color = "gray"
                    if (!progressTimer.running) {
                        progressTimer.start()
                    }
                    text = newText
                }

                function setTextError(newText) {
                    color = "red"
                    if (progressTimer.running) {
                        progressTimer.stop()
                    }
                    text = newText
                }

                Timer {
                    id: progressTimer
                    interval: 1500
                    running: true
                    repeat: true
                    onTriggered: {
                        progress.opacity = progress.opacity == 1.0 ? 0.1 : 1.0
                    }
                }
            }

            Image {
                id: settingButton
                anchors.top: progress.bottom
                anchors.topMargin: parent.height * 0.02
                anchors.horizontalCenter: logo.horizontalCenter
                width: Math.min(parent.width, parent.height) * 0.1
                height: width
                source: "../icon/settings.svg"
                smooth: true
                opacity: 0.0

                Behavior on opacity {
                    NumberAnimation {
                        duration: 500
                    }
                }

                MouseArea {
                    anchors.fill: parent
                    onClicked: {
                        console.log("openning settings")
                    }
                }
            }
        }

        /*
            MapView {
                id: mapView
                enabled: parent.flipped
                anchors.fill: parent
                showControls: leftMenu.state == "hidden"
                active: !leftMenu.visible

                onAction: {
                    handleAction(action)
                }

                onParentChanged: {
                    appWindow.actionCallback = getActionCallback()
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
                                             "name": itemName,
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
        }*/
    }
}
