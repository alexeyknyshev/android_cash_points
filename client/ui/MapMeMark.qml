import QtQuick 2.4
import QtLocation 5.3
import QtGraphicalEffects 1.0

MapQuickItem {
    id: item
    width: 128
    height: 128
    anchorPoint.x: width  * 0.5
    anchorPoint.y: height * 0.5
    sourceItem: Item {
        width: item.width
        height: item.height

        Image {
            anchors.fill: parent
            id: image
            source: "image://ico/marker.svg"
            sourceSize.width: item.width
            sourceSize.height: item.height
        }
        FastBlur {
            transparentBorder: true
            anchors.fill: parent
            source: image
            radius: 8
        }
    }
}
