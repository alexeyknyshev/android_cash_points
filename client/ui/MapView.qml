import QtQuick 2.4
import QtQuick.Controls 1.3
import QtLocation 5.3
import QtPositioning 5.3
import QtSensors 5.3
import QtGraphicalEffects 1.0

Item {
    id: topView
    width: 480
    height: 800
    visible: true

    focus: true

    property bool active: true
    property bool showZoomLevel: true
    property bool showControls: true

    property real contolsOpacity: showControls ? 1.0 : 0.0

    signal clicked()
    signal menuClicked()

    Keys.onEscapePressed: Qt.quit()
    Keys.onPressed: {
        if (event.key === Qt.Key_Plus) {
            map.zoom(map.zoomLevel + 1)
        } else if (event.key === Qt.Key_Minus) {
            map.zoom(map.zoomLevel - 1)
        }
    }

    Plugin {
        id: mapPlugin
        name: "osm"
    }

    function aboutToClose() {
        if (holdDialog.visible) {
            holdDialog.hide()
            return false
        }
        return true
    }

    PositionSource {
        id: positionSource
        name: "osm"
        updateInterval: 200
        preferredPositioningMethods: PositionSource.AllPositioningMethods
        onPositionChanged: {
            console.log("position changed")
            map.showMyPos()
            stop()
        }

        function getAvgZoomLevel(scale) {
            return map.minimumZoomLevel + (map.maximumZoomLevel - map.minimumZoomLevel) * scale
        }

        onActiveChanged: {
            console.log("active changed: " + active.toString())
        }

        function debugPrintSupportedPositioningMethods() {
            if (supportedPositioningMethods == PositionSource.AllPositioningMethods) {
                console.log("All pos methods")
                return
            }
            if (supportedPositioningMethods & PositionSource.SatellitePositioningMethods) {
                console.log("Sat pos methods")
            }
            if (supportedPositioningMethods & PositionSource.NonSatellitePositioningMethods) {
                console.log("Network pos methods")
            }
            if (supportedPositioningMethods & PositionSource.NoPositioningMethods) {
                console.log("No pos methods")
            }
        }
    }

    EnableLocServiceDialog {
        id: enableLocServiceDialog
        onVisibilityChanged: {
            console.log("dialog showed")
        }
        onRejected: {
            findMeButtonRotationAnim.stop()
            findMeButtonRotationAnimBack.start()
        }
        onAccepted: {
            locationService.enabled = true
        }
    }

    MapHoldDialog
    {
        id: holdDialog
        visible: false
        z: parent.z + 1
    }

    Map {
        id: map
        anchors.fill: parent
        plugin: mapPlugin;
        center {
            id: center
            latitude: coordLatitude
            longitude: coordLongitude
        }
        zoomLevel: 13
        maximumZoomLevel: 50
/*
        MapItemView {
//            model:
            delegate: cashPointDelegate
        }

        Component {
            id: cashPointDelegate


        }
*/
        property MapQuickItem me: null

        property bool panActive: false
        property real panLastX: 0
        property real panLastY: 0

        property real targetZoomLevel: 13
        property real coordLatitude: 55.7522200
        property real coordLongitude: 37.6155600

        gesture.flickDeceleration: 3000
        gesture.enabled: false
        gesture.activeGestures: MapGestureArea.FlickGesture | MapGestureArea.PanGesture

        property bool zooming: false

        ParallelAnimation {
            id: mapMoveAnim

            PropertyAnimation {
                id: zoomLevelAnim
                target: map
                property: "zoomLevel"
                to: map.targetZoomLevel
                easing.type: Easing.InOutQuad
                duration: 300
            }

            PropertyAnimation {
                id: latitudeAnim
                target: map
                property: "center.latitude"
                to: map.coordLatitude
                easing.type: Easing.InOutQuad
                duration: 300
            }

            PropertyAnimation {
                id: longitudeAnim
                target: map
                property: "center.longitude"
                to: map.coordLongitude
                easing.type: Easing.InOutQuad
                duration: 300
            }
        }


        function zoom(zoomFactor) {
            if (zoomFactor > maximumZoomLevel) {
                targetZoomLevel = maximumZoomLevel
            } else if (zoomFactor < minimumZoomLevel) {
                targetZoomLevel = minimumZoomLevel
            } else {
                targetZoomLevel = zoomFactor
            }

            var coord = map.toCoordinate(Qt.point(width * 0.5, height * 0.5))
            map.coordLatitude = coord.latitude
            map.coordLongitude = coord.longitude

            if (mapMoveAnim.running) {
                mapMoveAnim.stop()
            }

            mapMoveAnim.start()
        }

        function moveToCoord(coord, zoom) {
            console.warn("Coordinate:", coord.latitude, coord.longitude);
            map.coordLatitude = coord.latitude
            map.coordLongitude = coord.longitude
            map.targetZoomLevel = zoom
            if (mapMoveAnim.running) {
                mapMoveAnim.stop();
            }
            mapMoveAnim.start()
        }

        function showMyPos() {
            if (positionSource.valid) {
                if (positionSource.position.latitudeValid && positionSource.position.longitudeValid) {
                    var expectedZoomLevel = positionSource.getAvgZoomLevel(0.9)
                    if (expectedZoomLevel < map.zoomLevel) {
                        expectedZoomLevel = map.zoomLevel
                    }
                    moveToCoord(positionSource.position.coordinate, expectedZoomLevel)
                    if (me == null) {
                        var mapMeMarkComponent = Qt.createComponent("MapMeMark.qml")
                        if (mapMeMarkComponent.status === Component.Ready) {
                            me = mapMeMarkComponent.createObject(map, { width: Math.min(map.width, map.height) * 0.075,
                                                                        height: Math.min(map.width, map.height) * 0.075 })
                        }
                    }
                    me.coordinate = positionSource.position.coordinate
                    map.addMapItem(me)
                    return true
                }
            }
            return false
        }

        function findMe() {
            if (!locationService.enabled) {
                enableLocServiceDialog.open()
                return false
            }

            positionSource.stop();
            showMyPos()
            positionSource.start()
            return true
        }

        PinchArea {
            anchors.fill: parent

            property real oldZoom: 13

            onParentChanged: oldZoom = parent.zoomLevel

            onPinchStarted: {
                if (topView.active) {
                    console.log("pinch started")
                    oldZoom = map.zoomLevel
                }
            }

            onPinchUpdated: {
                if (topView.active) {
                    console.log("pinch")
                    console.log("scale: " + pinch.scale)
                    map.zoomLevel = oldZoom + Math.log(pinch.scale) / Math.log(2)
                }
            }

            MouseArea {
                id: mapMouseArea
                anchors.fill: parent
                z: parent.z + 1

                onPressed: {
                    map.panLastX = mouseX
                    map.panLastY = mouseY
                }

                onClicked: {
                    topView.clicked()

                    if (!topView.active)
                        return

                    if (parent.zooming)
                    {
                        parent.zooming = false
                        return
                    }

                    if (!holdDialog.visible) {
                        console.log(mouseX, mouseY)
                        console.log(map.toCoordinate(Qt.point(mouseX, mouseY)))
                        console.log(map.center)
//                        console.log("Map clicked!")
                    } else {
                        holdDialog.hide()
                    }
                }
//                onPressAndHold: {
//                    if (!map.zooming) {
//                        console.log("Hoooold...")
//                        holdDialog.show()
//                    }
//                }
                onDoubleClicked: {
                    if (!topView.active) {
                        return
                    }
//                    console.log("double click!")

                    var coord = map.toCoordinate(Qt.point(mouseX, mouseY))

                    map.moveToCoord(coord, map.zoomLevel + 1)
                }
                onPositionChanged: {
                    if (!topView.active)
                        return

                    var deltaX = mouseX - map.panLastX
                    var deltaY = mouseY - map.panLastY

                    var coord = map.toCoordinate(Qt.point(width / 2 - deltaX, height / 2 - deltaY))

                    map.center = coord

                    map.panLastX = mouseX
                    map.panLastY = mouseY

//                    console.log("mouse moved")
                }

                onReleased: {
//                    console.log("mouse released")
                }
            }
        }



        Image {
            z: parent.z + 1
            id: zoomOutButton
            source: "image://ico/zoom_out.svg"
            sourceSize: Qt.size(width, height)
            anchors.right: parent.right
            anchors.verticalCenter: parent.verticalCenter
            anchors.rightMargin: height * 0.25
            height: Math.min(Math.max(parent.width, parent.height) * 0.1, 160)
            width: height

            visible: opacity > 0
            opacity: topView.contolsOpacity
            Behavior on opacity {
                NumberAnimation {
                    duration: 300
                }
            }

            Behavior on scale {
                NumberAnimation { duration: 100 }
            }

            Rectangle {
                anchors.centerIn: parent
                width: parent.width * 0.3
                height: parent.height * 0.05
                radius: height * 0.1
            }

            MouseArea {
                anchors.fill: parent
                onClicked: if (topView.active) map.zoom(map.zoomLevel - 1)
                onPressed: if (topView.active) parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }

        }


        Image {
            z: parent.z + 1
            id: zoomInButton
            source: "image://ico/zoom_in.svg"
            sourceSize: Qt.size(width, height)
            anchors.top: zoomOutButton.bottom
            anchors.topMargin: height * 0.25
            anchors.right: parent.right
            anchors.rightMargin: height * 0.25
            height: Math.min(Math.max(parent.width, parent.height) * 0.1, 160)
            width: height

            visible: opacity > 0
            opacity: topView.contolsOpacity
            Behavior on opacity {
                NumberAnimation {
                    duration: 300
                }
            }

            Behavior on scale {
                NumberAnimation {
                    duration: 100
                }
            }

            MouseArea {
                anchors.fill: parent
                onClicked: if (topView.active) map.zoom(map.zoomLevel + 1)
                onPressed: if (topView.active) parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }
        }

//        InnerShadow {
//            anchors.fill: zoomOutButton
//            radius: 128.0
//                    samples: 16
//                    horizontalOffset: -10
//                    verticalOffset: -10
//                    color: "#b0000000"
//                    source: zoomOutButton
//        }

        Image {
            id: findMeButton
            width: height
            source: "image://ico/aim.svg"
            sourceSize: Qt.size(width, height)
            z: parent.z + 1
            anchors.left: parent.left
            anchors.leftMargin: height * 0.25
            anchors.verticalCenter: parent.verticalCenter
            anchors.margins: 1
            height: Math.min(Math.max(parent.width, parent.height) * 0.1, 160)

            visible: opacity > 0
            opacity: topView.contolsOpacity
            Behavior on opacity {
                NumberAnimation {
                    duration: 300
                }
            }

            property bool activated: false

            SequentialAnimation {
                id: findMeButtonRotationAnim
                running: false
                loops: 3
                PropertyAnimation {
                    target: findMeButton
                    property: "rotation"
                    to: 180
                    duration: 1500
                    easing.type: Easing.InOutCubic
                }
                RotationAnimation {
                    target: findMeButton
                    property: "rotation"
                    to: 0
                    duration: 1500
                    easing.type: Easing.InOutCubic
                }
                onStopped: {
                    if (!positionSource.valid) {
                        enableLocServiceDialog.open()
                    }
                }
            }

            RotationAnimation {
                id: findMeButtonRotationAnimBack
                running: false
                target: findMeButton
                property: "rotation"
                to: 0
                duration: 1500 * (findMeButton.rotation / 180)
            }

            Behavior on scale {
                NumberAnimation { duration: 100 }
            }

            MouseArea {
                anchors.fill: parent
                onClicked: {
                    if (!topView.active) {
                        return
                    }

                    var animate = map.findMe()
                    if (animate) {
                       findMeButtonRotationAnim.start()
                    }
                }
                onPressed: parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }
        }
    }

    Rectangle {
        id: searchLineEditContainer

        z: parent.z + 2

        anchors.top: parent.top
        anchors.topMargin: anchors.leftMargin
        anchors.left: parent.left
        anchors.leftMargin: parent.width * 0.02
        anchors.right: parent.right
        anchors.rightMargin: anchors.leftMargin
        anchors.bottomMargin: anchors.leftMargin

        radius: height / 20
        height: searchLineEdit.contentHeight * 2

//        color: "lightgray"

        TextInput {
            id: searchLineEdit

            z: parent.z + 1

            anchors.top: parent.top
            anchors.topMargin: searchLineEdit.contentHeight * 0.5
            anchors.left: menuButton.right
            anchors.leftMargin: searchLineEdit.contentHeight * 0.5
            anchors.right: clearButton.left
            anchors.rightMargin: searchLineEdit.contentHeight * 0.5
            anchors.bottomMargin: searchLineEdit.contentHeight * 0.5
            anchors.bottom: parent.bottom

            echoMode: TextInput.Normal

            font.pixelSize: Math.max(topView.height, topView.width) / (15 * 2) > 18 ?
                            Math.max(topView.height, topView.width) / (15 * 2) : 18

            property bool isUserTextShowed: false
            property string placeHolderText: qsTr("Искать")
            property string userText: ""

            wrapMode: Text.NoWrap

            Component.onCompleted: {
                text = placeHolderText
                color = "lightgray"
            }

            onFocusChanged: {
                if (searchLineEdit.focus) {
                    text = userText
                    color = "black"
                    isUserTextShowed = true
                } else {
                    userText = text
                    if (userText == "") {
                        text = placeHolderText
                        color = "lightgray"
                        isUserTextShowed = false
                    }
                }
            }

            function setFirstLetterUpper(upper)
            {
                if (upper && text != "") {
                    text = text.charAt(0).toUpperCase() + text.slice(1)
                }
            }

            onDisplayTextChanged: {
                if (displayText === "" || displayText === placeHolderText) {
                    //bankListModel.setFilter("")
                } else {
                    console.log("text changed")
                    //bankListModel.setFilter(displayText)
                }
            }
        } // TextInput

        Rectangle {
            id: menuButton
            z: parent.z + 1

            anchors.left: parent.left
            anchors.leftMargin: searchLineEdit.contentHeight * 0.2
            anchors.bottom: parent.bottom
            anchors.bottomMargin: searchLineEdit.contentHeight * 0.3
            anchors.top: parent.top
            anchors.topMargin: searchLineEdit.contentHeight * 0.3
            width: height

            Rectangle {
                id: menuButtonTopRect
                anchors.top: parent.top
                anchors.topMargin: parent.height * 2 / 9
                anchors.left: parent.left
                anchors.leftMargin: parent.width * 0.2
                height: parent.height / 9
                width: parent.width * 0.6
                color: "gray"
            }

            Rectangle {
                id: menuButtonCenterRect
                anchors.top: menuButtonTopRect.bottom
                anchors.topMargin: parent.height / 9
                anchors.left: parent.left
                anchors.leftMargin: parent.width * 0.2
                height: menuButtonTopRect.height
                width: parent.width * 0.6
                color: "gray"
            }

            Rectangle {
                anchors.top: menuButtonCenterRect.bottom
                anchors.topMargin: parent.height / 9
                anchors.left: parent.left
                anchors.leftMargin: parent.width * 0.2
                height: menuButtonTopRect.height
                width: parent.width * 0.6
                color: "gray"
            }

            MouseArea {
                anchors.fill: parent
                onClicked: {
                    console.log("menu clicked")
                    topView.menuClicked()
                }
            }
        }

        Rectangle {
            id: clearButton
            color: "transparent"
            anchors.right: parent.right
            anchors.top: parent.top
            anchors.topMargin: searchLineEdit.topMargin
            anchors.bottom: parent.bottom
            anchors.bottomMargin: searchLineEdit.bottomMargin
            width: height
            opacity: searchLineEdit.isUserTextShowed && searchLineEdit.displayText != "" ? 1.0 : 0.0

            Behavior on opacity {
                NumberAnimation { duration: 100 }
            }

            Rectangle {
                anchors.centerIn: parent
                height: parent.width * 0.05
                width: parent.width * 0.5
                color: "gray"
                rotation: 45
            }
            Rectangle {
                anchors.centerIn: parent
                height: parent.width * 0.05
                width: parent.width * 0.5
                color: "gray"
                rotation: -45
            }

            states: State {
                name: "pressed"
                PropertyChanges {
                    target: clearButton
                    color: "lightgray"
                }
            }

            MouseArea {
                anchors.fill: parent
                onClicked: {
                    searchLineEdit.userText = ""
                    searchLineEdit.text = ""
                }

                onHoveredChanged: {
                    if (containsMouse) {
                        parent.state = "pressed"
                    } else {
                        parent.state = ""
                    }
                }
            }
        }
    }

    RectangularGlow {
        z: searchLineEditContainer.z - 1
        anchors.fill: searchLineEditContainer
        glowRadius: searchLineEditContainer.height / 5
        spread: 0.3
        color: "#11000055"
        cornerRadius: glowRadius
    }

    Rectangle {
        width: Math.min(parent.width, parent.height) * 0.25
        height: Math.max(parent.width, parent.height) * 0.05
        radius: Math.min(width, height) * 0.1
        color: "white"
        visible: parent.showZoomLevel

        anchors.right: parent.right
        anchors.bottom: parent.bottom
        anchors.margins: 3

        Label {
            anchors.fill: parent
            text: "zoom: " + map.zoomLevel.toPrecision(6)
            color: "steelblue"
            font.bold: true
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
        }
    }
}

