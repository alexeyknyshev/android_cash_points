import QtQuick 2.0
import QtGraphicalEffects 1.0
import QtQuick.Window 2.2

MapView {
    id: mapView
    anchors.fill: parent
    showControls: leftMenu.state === "hidden"
    active: !leftMenu.visible

    property var externalAction

    onParentChanged: {
        appWindow.actionCallback = getActionCallback()
    }

    onAction: {
        if (externalAction && externalAction.actionCallback) {
            externalAction.actionCallback(action)
        }
    }

    LeftMenu {
        id: leftMenu
        x: 0
        z: mapView.z + 10
        anchors.top: parent.top
        anchors.bottom: parent.bottom
        width: parent.width * (Screen.primaryOrientation == Screen.orientation ? 0.6 : 0.4)

        onAction: {
            if (act.type && act.type === "moveToCoord") {
                mapView.moveToCoord(act.coord, act.zoom ? act.zoom : mapView.getMapZoom())
            } else {
                mapView.action(act)
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
            action({
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
        action({
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
