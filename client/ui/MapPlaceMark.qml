import QtQuick 2.4
import QtLocation 5.3

MapQuickItem {
    id: item
    width: 128
    height: 128
    anchorPoint.x: width  * 0.5
    anchorPoint.y: height

    property alias source: marker.source

    sourceItem: Item {
        width: item.width
        height: item.height

        Image {
            id: marker
            anchors.fill: parent
            sourceSize.width: item.width
            sourceSize.height: item.height
        }
    }
}

