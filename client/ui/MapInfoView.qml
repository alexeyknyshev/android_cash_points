import QtQuick 2.0

Rectangle {
    id: topItem

    property real hiddenY: y

    function isHidden() {
        return state == "hidden"
    }

    function hide() {
        if (state != "hidden") {
            state = "hidden"
            return true
        }
        return false
    }

    function isShown() {
        return state == ""
    }

    function show() {
        if (state != "") {
            state = ""
            return true
        }
    }

    function isShownFullSceen() {
        return state == "fullscreen"
    }

    function showFullscreen() {
        if (state != "fullscreen") {
            state = "fullscreen"
            return true
        }    }

    states: [
        State {
            name: "hidden"
            PropertyChanges {
                target: topItem
                y: hiddenY
            }
        },
        State {
            name: "fullscreen"
            PropertyChanges {
                target: topItem
                height: parent.height
                y: 0
            }
        }
    ]

    transitions: [
        Transition {
            from: ""
            to: "hidden"
            PropertyAnimation {
                duration: 500
                easing.type: Easing.InOutQuart
                properties: "y"
            }
            onRunningChanged: {
                console.warn("MapInfoView: started")
                if (!running) {
                    visible = false
                }
            }
        },
        Transition {
            from: "hidden"
            to: ""
            PropertyAnimation {
                duration: 500
                easing.type: Easing.InOutQuart
                properties: "y"
            }
            onRunningChanged: {
                if (running) {
                    visible = true
                }
            }
        }
    ]

    Row {
        anchors.fill: parent
        anchors.margins: parent.height * 0.05
        Rectangle {
            color: "blue"
            height: parent.height * 0.9
            width: height
        }
    }
}

