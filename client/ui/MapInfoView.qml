import QtQuick 2.0

Rectangle {
    id: topItem

    y: parent.height - height
    property real shownY: parent.height

    function isHidden() {
        return state == ""
    }

    function hide() {
        if (state != "") {
            state = ""
        }
    }

    function isShown() {
        return state == "shown"
    }

    function show() {
        if (state != "shown") {
            state = "shown"
        }
    }

    function isShownFullSceen() {
        return state == "fullscreen"
    }

    function showFullscreen() {
        if (state != "fullscreen") {
            state = "fullscreen"
        }
    }

    states: [
        State {
            name: "shown"
            PropertyChanges {
                target: topItem
                y: shownY
                height: parent.height - shownY
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
            from: "shown"
            to: ""
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
            from: ""
            to: "shown"
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
        },
        Transition {
            from: "shown"
            to: "fullscreen"
            PropertyAnimation {
                duration: 500
                easing.type: Easing.InOutQuart
                properties: "y, height"
            }
        },
        Transition {
            from: "fullscreen"
            to: ""
            PropertyAnimation {
                duration: 500
                easing.type: Easing.InOutQuart
                properties: "y, height"
            }
        }
    ]

    /*Row {
        anchors.fill: parent
        anchors.margins: parent.height * 0.05
        Rectangle {
            color: "blue"
            height: parent.height * 0.9
            width: height
        }
    }*/
}

