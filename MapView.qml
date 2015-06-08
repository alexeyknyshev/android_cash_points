import QtQuick 2.4
import QtQuick.Controls 1.3
import QtLocation 5.3
import QtPositioning 5.3
import QtSensors 5.3

Item {
    width: 480
    height: 800
    visible: true

    focus: true

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
        updateInterval: 200
        preferredPositioningMethods: PositionSource.AllPositioningMethods
        property bool needUpdate: false
        property MapCircle circle
        onPositionChanged: {
            if (position.coordinate.isValid && needUpdate) {
                needUpdate = false
                map.moveToCoord(position.coordinate)
                map.zoom(map.minimumZoomLevel + (map.maximumZoomLevel - map.minimumZoomLevel) * 0.9)

                    if (circle == null) {
                        circle = Qt.createQmlObject('import QtLocation 5.3; MapCircle {}', map)
                    }
                    circle.center = position.coordinate
                    circle.radius = 15.0
                    circle.color = 'green'
                    circle.border.width = 3
                    map.addMapItem(circle)

               findMeButton.scaleAnimated = false
            }
        }

        function forceUpdate() {
            needUpdate = true
            update()
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

        PropertyAnimation {
            id: zoomLevelAnim
            target: map
            property: "zoomLevel"
            to: map.targetZoomLevel
            duration: 300
            easing.type: Easing.InOutQuad
//            onStarted: console.log("started")
//            onStopped: console.log("stopped")
        }


        function zoom(zoomFactor) {
            if (zoomFactor > maximumZoomLevel) {
                targetZoomLevel = maximumZoomLevel
            } else if (zoomFactor < minimumZoomLevel) {
                targetZoomLevel = minimumZoomLevel
            } else {
                targetZoomLevel = zoomFactor
            }

            if (zoomLevelAnim.running) {
                zoomLevelAnim.stop()
            }

            zoomLevelAnim.start()
        }

        function moveToCoord(coord) {
            console.log("Coordinate:", coord.latitude, coord.longitude);
            map.center = coord
        }

        function findMe() {
            if (positionSource.supportedPositioningMethods ===
                    PositionSource.NoPositioningMethods)
            {
                console.error("positioning methods are unsupported!")
                return
            }

            if (!positionSource.active) {
                positionSource.start()
            }

            positionSource.forceUpdate()
        }

        PinchArea {
            anchors.fill: parent

            property real oldZoom: 13

            onParentChanged: oldZoom = parent.zoomLevel

            onPinchStarted: {
                console.log("pinch started")
                oldZoom = map.zoomLevel
            }

            onPinchUpdated: {
                console.log("pinch")
                console.log("scale: " + pinch.scale)
                map.zoomLevel = oldZoom + Math.log(pinch.scale) / Math.log(2)
            }

            MouseArea {
                id: mapMouseArea
                anchors.fill: parent
                z: parent.z + 1

                onClicked: {
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
//                    console.log("double click!")

                    var coord = map.toCoordinate(Qt.point(mouseX, mouseY))

                    // prevent flicker near center
                    if (Math.abs(mouseX - width  / 2) / width  > 0.05 ||
                        Math.abs(mouseY - height / 2) / height > 0.05)
                    {
                        map.moveToCoord(coord)
                    }

                    map.zoom(map.zoomLevel + 1)
                }
                onPositionChanged: {
                    if (!map.panActive) {
                        map.panActive = true
                        map.panLastX = mouseX
                        map.panLastY = mouseY
                        return
                    }

                    var deltaX = mouseX - map.panLastX
                    var deltaY = mouseY - map.panLastY

//                    console.log("deltaX: " + deltaX)
//                    console.log("deltaY: " + deltaY)

                    var coord = map.toCoordinate(Qt.point(width / 2 - deltaX, height / 2 - deltaY))

                    map.center = coord

                    map.panLastX = mouseX
                    map.panLastY = mouseY

//                    console.log("mouse moved")
                }

                onReleased: {
                    map.panActive = false
//                    console.log("mouse released")
                }
            }
        }



        Image {
            z: parent.z + 1
            id: zoomOutButton
            anchors.right: parent.right
            anchors.verticalCenter: parent.verticalCenter
            anchors.margins: 1
            height: Math.min(Math.max(parent.width, parent.height) * 0.125, 160)
            width: height
            source: "ico/ico/gtk-zoom-out.png"
            opacity: 0.95

            MouseArea {
                anchors.fill: parent
                onClicked: map.zoom(map.zoomLevel - 1)
                onPressed: parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }

        }

        Image {
            z: parent.z + 1
            id: zoomInButton
            anchors.top: zoomOutButton.bottom
            anchors.right: parent.right
            anchors.margins: 1
            height: Math.min(Math.max(parent.width, parent.height) * 0.125, 160)
            width: height
            source: "ico/ico/gtk-zoom-in.png"
            opacity: 0.95

            MouseArea {
                anchors.fill: parent
                onClicked: map.zoom(map.zoomLevel + 1)
                onPressed: parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }
        }

        Image {
                z: parent.z + 1
                anchors.left: parent.left
                anchors.verticalCenter: parent.verticalCenter
                anchors.margins: 1
                height: Math.min(Math.max(parent.width, parent.height) * 0.125, 160)
                width: height
                id: findMeButton
                source: "ico/ico/zoom-fit-best.png"

                property bool scaleAnimated: false

                states: [
                    State {
                        name: "animated_out"
                        PropertyChanges {
                            target: findMeButton
                            scale: 0.8
                        }
                        onCompleted: findMeButton.state = "animated_in"
                    },
                    State {
                        name: "animated_in"
                        PropertyChanges {
                            target: findMeButton
                            scale: 1.0
                        }
                        onCompleted: {
                            if (findMeButton.scaleAnimated) {
                                findMeButton.state = "animated_out"
                            } else {
                                findMeButton.state = ""
                            }
                        }
                    }
                ]

                transitions: Transition {
                    ScaleAnimator {
                        duration: 2000
                    }
                }

                MouseArea {
                    anchors.fill: parent
                    onClicked: {
                        map.findMe()
                        if (parent.state === "") {
                            parent.scaleAnimated = true
                            parent.state = "animated_out"
                        }
                    }
                    onPressed: parent.scale = 0.9
                    onReleased: parent.scale = 1.0
                }
        }
    }
}

