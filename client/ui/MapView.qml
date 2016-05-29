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
    property bool showZoomLevel: false
    property bool showControls: true

    property real cashpointZoom: 15

    property real contolsOpacity: showControls ? 1.0 : 0.0

    signal clicked()
    signal menuClicked()

    signal action(var action)

    function clamp(val, min, max) {
        return Math.max(min, Math.min(max, val))
    }

    Keys.onEscapePressed: Qt.quit()
    Keys.onPressed: {
        if (event.key === Qt.Key_Plus) {
            map.moveToCoord(null, map.zoomLevel + 1)
        } else if (event.key === Qt.Key_Minus) {
            map.moveToCoord(null, map.zoomLevel - 1)
        }
    }

    function getActionCallback() {
        return function(action) {
            if (action.type === "undo" && infoView.state !== "hidden") {
                infoView.hide()
                return false
            }
            return true
        }
    }

    function cashpointTypePrintable(type) {
        if (type === "atm") {
            return qsTr("Банкомат")
        } else if (type === "office" || type === "branch") {
            return qsTr("Офис")
        } else if (type === "cash") {
            return qsTr("Касса")
        }
        return ""
    }

    function currencyTypePrintable(cp) {
        var text = ""
        if (cp.rub) {
            text += qsTr("рубли")
        }
        if (cp.eur) {
            if (text.length > 0) {
                text += ", "
            }
            text += qsTr("доллары")
        }
        if (cp.usd) {
            if (text.length > 0) {
                text += ", "
            }
            text += qsTr("евро")
        }
        return text
    }

    function getVisiableAreaRadius() {
        var radius = 0//locationService.getGeoRegionRadius(map.visibleRegion)
        if (radius <= 10.0) {
            console.log(map.center)
            console.log(map.toCoordinate(Qt.point(0.0, 0.0)))
            radius = locationService.getGeoRegionRadiusEstimate(map.center, map.toCoordinate(Qt.point(0.0, 0.0)))
        }
        return radius
    }

    function getMapCenter() {
        return map.center
    }

    function getMapZoom() {
        return map.zoomLevel
    }

    onEnabledChanged: {
        if (enabled) {
            map.invalidate()
        }
    }

    Plugin {
        id: mapPlugin
        name: "osm"
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
        maximumZoomLevel: 18.9

        onParentChanged: {
            var lastPos = JSON.parse(cashpointModel.getLastGeoPos())
            coordLatitude = lastPos.latitude
            coordLongitude = lastPos.longitude
            targetZoomLevel = lastPos.zoom
            _moveToCoord(QtPositioning.coordinate(lastPos.latitude, lastPos.longitude), lastPos.zoom)
        }

        function targetCoord() {
            return QtPositioning.coordinate(coordLatitude, coordLongitude)
        }

        MapItemView {
            model: cashpointModel
            delegate: MapQuickItem {
                id: item
                anchorPoint.x: width * 0.5
                anchorPoint.y: model.cp_type === "cluster" ? height * 0.5 : height
                height: Math.min(map.width, map.height) * 0.1
                width: Math.min(map.width, map.height) * 0.1
                coordinate {
                    longitude: model.cp_coord_lon
                    latitude: model.cp_coord_lat
                }

                property var pointId: model.cp_id
                property var pointBankId: model.cp_bank_id
                property var pointType: model.cp_type
                property var pointName: model.cp_name
                property var pointAddress: model.cp_address

                sourceItem: Item {
                        property real cpSizeLen: model.cp_size.toString().length
                        property real multiplier: model.cp_type === "cluster" ? (cpSizeLen > 3 ? 1.0 + (cpSizeLen - 3) * 0.25 : 1.0) : 1.0
                        width: item.width * (multiplier > 1.0 ? multiplier : 1.0)
                        height: item.height * (multiplier > 1.0 ? multiplier : 1.0)

                        onMultiplierChanged: {
                            console.log(multiplier)
                        }

                        opacity: topView.active ? 1.0 : 0.0
                        visible: opacity > 0.0
                        Behavior on opacity {
                            NumberAnimation { duration: 400 }
                        }

                        Image {
                            id: marker
                            anchors.centerIn: parent
                            sourceSize.width: parent.width
                            sourceSize.height: parent.height
                            source: model.cp_type === "cluster" ? "image://ico/cluster.svg" : "image://ico/place.svg"

                            Text {
                                visible: model.cp_type === "cluster"
                                anchors.centerIn: parent
                                verticalAlignment: Text.AlignVCenter
                                horizontalAlignment: Text.AlignHCenter
                                text: model.cp_size ? model.cp_size : 0
                            }
                        }

                        Image {
                            id: logo
                            visible: model.cp_type !== "cluster"
                            anchors.fill: parent
                            anchors.leftMargin: parent.width * 0.22
                            anchors.rightMargin: parent.width * 0.22
                            anchors.topMargin: parent.width * 0.07
                            anchors.bottomMargin: parent.width * 0.36
                            sourceSize.width: parent.width * 0.8
                            sourceSize.height: parent.width * 0.8
                            source: model.cp_type === "cluster"
                                    ? "image://empty/"
                                    : "image://ico/bank/" + model.cp_bank_id
                        }
                    }
            }
        }

        property MapQuickItem me: null
        property MapQuickItem mark: null

        property bool panActive: false

        property real panLastX: 0
        property real panLastY: 0

        property real panStartX: 0
        property real panStartY: 0

        //property Timer invalidateTime: new Date()

        function getPanDistanceSqr() {
            var deltaX = panLastX - panStartX
            var deltaY = panLastY - panStartY
            return deltaX * deltaX + deltaY * deltaY
        }

        function invalidate() {
            var visiableRadius = getVisiableAreaRadius()
            console.warn("visiable radius: " + visiableRadius)

            var type = ""
            var zoom = targetZoomLevel
            if (zoom > 16) {
                type = "radius"
            } else {
                zoom--
                type = "cluster"
            }

            var topLeft = map.toCoordinate(Qt.point(0.0, 0.0))
            var botRight = map.toCoordinate(Qt.point(map.width, map.height))

            var json = {
                "type": type,
                "radius": visiableRadius,
                "longitude": coordLongitude,
                "latitude": coordLatitude,
                "zoom": zoom,
                "topLeft": {
                    "longitude": topLeft.longitude,
                    "latitude": topLeft.latitude,
                },
                "bottomRight": {
                    "longitude": botRight.longitude,
                    "latitude": botRight.latitude
                },
            }

            if (searchLineEdit.searchCandidate) {
                console.log("searchCandidate: " + JSON.stringify(searchLineEdit.searchCandidate))
                json.filter = searchLineEdit.searchCandidate.filter ? searchLineEdit.searchCandidate.filter : {}
            } else {
                var filterJson = searchEngine.getMineBanksFilter()
                json.filter = JSON.parse(filterJson)
            }

            if (json.filter.bank_id && searchEngine.showPartnerBanks) {
                var partners = bankListModel.getPartnerBanks(json.filter.bank_id)
                for (var i = 0; i < partners.length; i++) {
                    if (json.filter.bank_id.indexOf(partners[i]) === -1) {
                        json.filter.bank_id.push(partners[i])
                    }
                }
            }

            cashpointModel.setFilter(JSON.stringify(json))
        }

        function mapItemsAtScenePos(x, y) {
            var result = []
            for (var i = 0; i < mapItems.length; i++) {
                //console.log(map.mapItems[i].x + " " + map.mapItems[i].y)
                var itemHalfW = mapItems[i].width * 0.5
                var itemHalfH = mapItems[i].height * 0.5

                var itemCenterX = mapItems[i].x + itemHalfW
                var itemCenterY = mapItems[i].y + itemHalfH

                var itemSqrRadius = itemHalfW * itemHalfW + itemHalfH * itemHalfH

                var deltaX = itemCenterX - x
                var deltaY = itemCenterY - y

                var sqrDist = deltaX * deltaX + deltaY * deltaY

                if (sqrDist < itemSqrRadius) {
                    //console.log(mapItems[i].pointId)
                    //console.log(mapItems[i].pointType)
                    result.push({
                                    "id": mapItems[i].pointId,
                                    "type": mapItems[i].pointType,
                                    "dist": sqrDist, // ! pixel dist (not meters)
                                    "longitude": mapItems[i].coordinate.longitude,
                                    "latitude": mapItems[i].coordinate.latitude,
                                    "bank": mapItems[i].pointBankId,
                                    "name": mapItems[i].pointName,
                                    "address": mapItems[i].pointAddress,
                                    "radiusSqr": itemHalfH * itemHalfH + itemHalfW * itemHalfW
                                })
                }
            }
            return result
        }

        property real targetZoomLevel: 13
        property real coordLatitude: 0.0//55.7522200
        property real coordLongitude: -30.0//37.6155600

        gesture.flickDeceleration: 3000
        gesture.enabled: false
        gesture.activeGestures: MapGestureArea.FlickGesture | MapGestureArea.PanGesture

        property bool zooming: false

        ParallelAnimation {
            id: mapMoveAnim

            PropertyAnimation {
                id: latitudeAnim
                target: map
                property: "center.latitude"
                to: map.coordLatitude
                easing.type: Easing.InOutSine
                duration: 500
            }

            PropertyAnimation {
                id: longitudeAnim
                target: map
                property: "center.longitude"
                to: map.coordLongitude
                easing.type: Easing.InOutSine
                duration: 500
            }

            PropertyAnimation {
                id: zoomLevelAnim
                target: map
                property: "zoomLevel"
                to: map.targetZoomLevel
                easing.type: Easing.InOutQuad
                duration: 300
            }

            onStopped: {
                map.invalidate()
            }
        }

        function addCashpoint(coord) {
            var act = {
                "type": "addCashpoint",
                "prevCoord": map.targetCoord(),
                "prevZoom": map.targetZoomLevel,
                "coord": coord,
                "do": function(act) {
                    map._moveToCoord(act.coord, Math.max(map.zoomLevel, cashpointZoom))
                    map._addCashpoint(act.coord)
                    //infoTabView.page = 0
                    infoView.show()
                    return true
                },
                "undo": function(act) {
                    map._moveToCoord(act.prevCoord, act.prevZoom)
                    map._hideCashpoint()
                    infoView.hide()
                    return true
                }
            }

            action(act)
        }

        function moveToCoord(coord, zoom, fromCoord, round) {
            var act = {
                "type": "moveToCoord",
                "prevCoord": fromCoord ? fromCoord : map.targetCoord(),
                "coord": coord,
                "zoom": zoom,
                "prevZoom": map.targetZoomLevel,
                "round": round === undefined ? true : round,
                "do": function(act) {
                    map._moveToCoord(act.coord, act.zoom, act.round)
                    return true
                },
                "undo": function(act) {
                    map._moveToCoord(act.prevCoord, act.prevZoom, act.round)
                    return true
                }
            }

            action(act)
        }

        function _moveToCoord(coord, zoom, round) {
            if (!coord) {
                coord = map.toCoordinate(Qt.point(width * 0.5, height * 0.5))
            }

            console.log("Coordinate:", coord.latitude, coord.longitude);
            map.coordLatitude = coord.latitude
            map.coordLongitude = coord.longitude
            if (zoom) {
                if (round) {
                    zoom = Math.round(zoom)
                }
                map.targetZoomLevel = clamp(zoom, minimumZoomLevel, maximumZoomLevel)
            }
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
                console.log("here")
                return false
            }

            positionSource.stop();
            var showed = showMyPos()
            positionSource.start()
            console.log(showed)
            return showed
        }

        function _addCashpoint(coord) {
            if (!map.mark) {
                var mapMarkComponent = Qt.createComponent("MapPlaceMark.qml")
                map.mark = mapMarkComponent.createObject(map, { width: Math.min(map.width, map.height) * 0.1,
                                                                height: Math.min(map.width, map.height) * 0.1 })
                map.mark.source = "image://ico/place_add.svg"
                map.mark.logo = "";
                map.addMapItem(map.mark)
            } else {
                map.mark.visible = true
            }

            map.mark.coordinate = coord
        }

        function _hideCashpoint() {
            if (map.mark) {
                map.mark.visible = false
            }
        }

        PinchArea {
            anchors.fill: parent

            property real oldZoom: 13

            onParentChanged: {
                oldZoom = parent.zoomLevel
            }

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
                    addPointTimer.stop()
                }
            }

            onPinchFinished: {
                map.targetZoomLevel = map.zoomLevel
                map.moveToCoord(map.center, map.zoomLevel, map.targetCoord(), false)
            }

            MouseArea {
                id: mapMouseArea
                anchors.fill: parent
                z: parent.z + 1

                property real totalPanDistanceSqr: 0.0
                property real panDistanceThreshold: Math.min(width, height) * 0.05

                Timer {
                    id: addPointTimer
                    interval: 2500
                    onTriggered: {
                        var panDistanceThresholdSqr = mapMouseArea.panDistanceThreshold * mapMouseArea.panDistanceThreshold
                        if (mapMouseArea.totalPanDistanceSqr < panDistanceThresholdSqr) {
                            console.log("ADD!")
                            if (!topView.active) {
                                return
                            }
                            var minSideLen = Math.min(map.width, map.height)
                            var minSideLenSqr = minSideLen * minSideLen
                            if (map.getPanDistanceSqr() > minSideLenSqr * 0.01) {
                                return
                            }

                            var coord = map.toCoordinate(Qt.point(mapMouseArea.mouseX, mapMouseArea.mouseY))

                            map.addCashpoint(coord)
                        }
                        mapMouseArea.totalPanDistanceSqr = 0.0
                    }
                }

                onPressed: {
                    //searchSuggestionsView.visible = false
                    searchLineEdit.focus = false
                    searchEngine.setFilter("")

                    map.panLastX = mouseX
                    map.panLastY = mouseY

                    map.panStartX = mouseX
                    map.panStartY = mouseY

                    totalPanDistanceSqr = 0.0
                    addPointTimer.start()
                }

                onClicked: {
                    topView.clicked()

                    if (!topView.active) {
                        return
                    }

                    if (parent.zooming)
                    {
                        parent.zooming = false
                        return
                    }

                    if (!holdDialog.visible) {

                    } else {
                        holdDialog.hide()
                    }

                    var items = map.mapItemsAtScenePos(mouseX, mouseY)
                    items.sort(function(a, b) {
                        return a.dist - b.dist
                    })

                    //for (var i = 0; i < items.length; i++) {
                    //    console.log(JSON.stringify(items[i]))
                    //}

                    if (items.length > 0) {
                        var data = items[0]

                        console.log(JSON.stringify(data))

                        if (data.type === "cluster" && data.dist < data.radiusSqr) {
                            var minSideLen = Math.min(map.width, map.height)
                            var minSideLenSqr = minSideLen * minSideLen
                            if (map.getPanDistanceSqr() < minSideLenSqr * 0.01) {
                                map.moveToCoord({
                                                    "longitude": data.longitude,
                                                    "latitude": data.latitude,
                                                },
                                                map.zoomLevel + 1)
                            }
                        }
                    }
                }
                onDoubleClicked: {
                    if (!topView.active) {
                        return
                    }
                    var coord = map.toCoordinate(Qt.point(mouseX, mouseY))
                    map.moveToCoord(coord, map.zoomLevel + 1)
                }

                onPositionChanged: {
                    //console.log("pos changed")
                    if (!topView.active) {
                        return
                    }

                    var deltaX = mouseX - map.panLastX
                    var deltaY = mouseY - map.panLastY

                    if (!mapMoveAnim.running) {
                        var coord = map.toCoordinate(Qt.point(width / 2 - deltaX, height / 2 - deltaY))

                        if (map.mark && map.visibleRegion) {
                            if (!map.visibleRegion.contains(map.mark.coordinate)) {
                                map.mark.visible = false
                            }
                        }

                        map.center = coord
                        map.coordLongitude = coord.longitude
                        map.coordLatitude = coord.latitude
                    } else {
                        map.panStartX = mouseX
                        map.panStartY = mouseY
                    }

                    mapMouseArea.totalPanDistanceSqr += deltaX * deltaX + deltaY * deltaY

                    map.panLastX = mouseX
                    map.panLastY = mouseY
                }
                onReleased: {
                    var minSideLen = Math.min(map.width, map.height)
                    var minSideLenSqr = minSideLen * minSideLen
                    if (map.getPanDistanceSqr() >= minSideLenSqr * 0.01) {
                        var panStartCoord = map.toCoordinate(Qt.point(mouseX, mouseY))
                        var currentCoord = QtPositioning.coordinate(map.coordLatitude, map.coordLongitude)
                        map.moveToCoord(currentCoord, map.zoomLevel, panStartCoord, false)
                    } else {
                        console.log(mouseX, mouseY)
                        var coord = map.toCoordinate(Qt.point(mouseX, mouseY))
                        console.log(coord.latitude + " " + coord.longitude)

                        var items = map.mapItemsAtScenePos(mouseX, mouseY)
                        items.sort(function(a, b) {
                            return a.dist - b.dist
                        })

                        //for (var i = 0; i < items.length; i++) {
                        //    console.log(JSON.stringify(items[i]))
                        //}

                        if (items.length > 0) {
                            var data = items[0]
                            console.log(JSON.stringify(data))

                            if (data.type !== "cluster") {
                                var act = {
                                    //"prevPage": infoTabView.page,
                                    "prevData": infoView.data,
                                    "data": data,
                                    "type": "showInfo",
                                    "do": function(act) {
                                        infoView.setInfoData(act.data)
                                        infoView.show()
                                        //infoTabView.page = 1
                                        return true
                                    },
                                    "undo": function(act) {
                                        infoView.setInfoData(act.prevData)
                                        if (act.prevPage) {
                                            //infoTabView.page = act.prevPage
                                        }
                                        infoView.hide()
                                        return true
                                    }
                                }

                                action(act)

//                                if (items[0].bank) {
//                                    var bankJsonData = bankListModel.getBankData(items[0].bank)
//                                    if (bankJsonData !== "") {
//                                        var bank = JSON.parse(bankJsonData)
//                                        text += bank.name
//                                    }
//                                }

//                                addToMapText.text = text
                            }
                        }
                    }

                    addPointTimer.stop()
                }
            }
        }


        Image {
            z: parent.z + 1
            id: zoomInButton
            source: "image://ico/zoom_in.svg"
            sourceSize: Qt.size(width, height)            
            anchors.right: parent.right
            anchors.verticalCenter: parent.verticalCenter
            anchors.rightMargin: height * 0.25
            height: Math.min(Math.max(parent.width, parent.height) * 0.1, Math.min(parent.width, parent.height) * 0.15)
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
                onClicked: map.moveToCoord(map.center, map.zoomLevel + 1)
                onPressed: if (topView.active) parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }
        }

        Image {
            z: parent.z + 1
            id: zoomOutButton
            source: "image://ico/zoom_out.svg"
            sourceSize: Qt.size(width, height)
            anchors.top: zoomInButton.bottom
            anchors.topMargin: height * 0.25
            anchors.right: parent.right
            anchors.rightMargin: zoomInButton.anchors.rightMargin
            height: zoomInButton.height
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

            Rectangle {
                anchors.centerIn: parent
                width: parent.width * 0.3
                height: parent.height * 0.05
                radius: height * 0.1
            }

            MouseArea {
                anchors.fill: parent
                onClicked: map.moveToCoord(map.center, map.zoomLevel - 1)
                onPressed: if (topView.active) parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }

        }

        Image {
            z: parent.z + 1
            id: showMineBanks
            source: active ? "image://ico/star.svg" : "image://ico/star_gray.svg"
            sourceSize: Qt.size(width, height)
            anchors.bottom: parent.bottom
            anchors.right: parent.right
            anchors.margins: zoomInButton.anchors.rightMargin
            height: zoomInButton.height
            width: height

            property bool active: false

            Behavior on scale {
                NumberAnimation {
                    duration: 100
                }
            }

            MouseArea {
                id: showMineBanksMA
                anchors.fill: parent
                onClicked: if (topView.active) parent.active = !parent.active
                onPressed: if (topView.active) parent.scale = 0.9
                onReleased: parent.scale = 1.0
            }

            onActiveChanged: {
                searchEngine.showOnlyMineBanks = active
                map.invalidate()
            }
        }

        Image {
            id: findMeButton
            source: "image://ico/aim.svg"
            sourceSize: Qt.size(width, height)
            z: parent.z + 1
            anchors.left: parent.left
            anchors.leftMargin: height * 0.25
            anchors.verticalCenter: parent.verticalCenter
            anchors.margins: 1
            height: zoomInButton.height
            width: height

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

                    var found = map.findMe()
                    if (!found) {
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
        //anchors.bottomMargin: anchors.leftMargin

        radius: searchLineEdit.contentHeight * 0.1
        height: menuButton.height + (topView.height * 0.08) * searchEngine.rowCount +
                searchSuggestionsView.anchors.margins * 2 * (searchSuggestionsView.visible ? 1.0 : 0.0)

//        color: "lightgray"
        Behavior on height {
            NumberAnimation {
                duration: 200
            }
        }

        TextInput {
            id: searchLineEdit

            z: parent.z + 1

            anchors.top: parent.top
            anchors.topMargin: searchLineEdit.contentHeight * 0.5
            anchors.left: menuButton.right
            anchors.leftMargin: searchLineEdit.contentHeight * 0.5
            anchors.right: clearButton.left
            anchors.rightMargin: searchLineEdit.contentHeight * 0.5
            //anchors.bottomMargin: searchLineEdit.contentHeight * 0.5
            //anchors.bottom: parent.bottom

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
                    searchEngine.setFilter(userText)
                } else {
                    userText = text
                    if (userText == "") {
                        text = placeHolderText
                        color = "lightgray"
                        isUserTextShowed = false
                    }
                    searchEngine.setFilter("")
                }
            }

            onAccepted: {
                acceptSearchCandidate()
            }

            property var searchCandidate: null
            function setSearchCandidate(candidate) {
                console.log("searchCandidate: " + JSON.stringify(candidate))
                searchCandidate = candidate
//                if (candidate.filter) {
//                    searchEngine.filterPatch = JSON.stringify(candidate.filter)
//                } else {
//                    searchEngine.filterPatch = ""
//                }

                searchLineEdit.focus = true
                if (candidate) {
                    searchLineEdit.text = candidate.name
                    if (candidate.type === "town") {
                        map.moveToCoord(QtPositioning.coordinate(candidate.latitude, candidate.longitude),
                                        candidate.zoom)
                    } else {
                        map.invalidate()
                    }
                } else {
                    searchLineEdit.userText = ""
                    searchLineEdit.text = ""
                    searchLineEdit.focus = false
                    map.invalidate()
                }
            }

            function acceptSearchCandidate() {
                var candidateJson = searchEngine.getCandidate()
                if (candidateJson !== "") {
                    var candidate = JSON.parse(candidateJson)
                    if (candidate) {
                        setSearchCandidate(candidate)
                    }
                }
                focus = false
            }

            function setFirstLetterUpper(upper)
            {
                if (upper && text != "") {
                    text = text.charAt(0).toUpperCase() + text.slice(1)
                }
            }

            onDisplayTextChanged: {
                if (displayText === "" || displayText === placeHolderText) {
                    searchEngine.setFilter("")
                } else {
                    console.log("text changed")
                    searchEngine.setFilter(displayText)
                }
            }
        } // TextInput

        Rectangle {
            id: menuButton
            z: parent.z + 1

            anchors.left: parent.left
//            anchors.leftMargin: searchLineEdit.contentHeight * 0.2
//            anchors.bottomMargin: searchLineEdit.contentHeight * 0.3
            anchors.top: parent.top
//            anchors.topMargin: searchLineEdit.contentHeight * 0.3
            width: height

            height: searchLineEdit.height * 2
            color: "transparent"

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
//            anchors.bottom: parent.bottom
//            anchors.bottomMargin: searchLineEdit.bottomMargin
            height: menuButton.height
            width: height
            visible: searchLineEdit.focus && searchLineEdit.isUserTextShowed && searchLineEdit.displayText != ""

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
                    // remove search candidate
                    searchLineEdit.setSearchCandidate()
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

        ListView {
            id: searchSuggestionsView
            model: searchEngine

            visible: searchEngine.rowCount > 0 && searchLineEdit.focus
            anchors.top: searchLineEdit.bottom
            anchors.left: parent.left
            anchors.right: parent.right
            height: (topView.height * 0.08) * searchEngine.rowCount

            interactive: false

            Behavior on height {
                NumberAnimation {
                    duration: 200
                }
            }

            Connections {
                target: searchEngine
                onRowCountChanged: {
                    console.log("search rows: " + count.toString())
                    searchSuggestionsView.visible = (count > 0)
                    searchSuggestionsView.height = (topView.height * 0.08) * count
                }
            }

            delegate: Rectangle {
                height: (topView.height * 0.08)
                width: parent.width
                color: "transparent"

                Image {
                    id: suggestIco
//                    anchors.margins: searchLineEdit.anchors.leftMargin
                    anchors.leftMargin: menuButton.anchors.leftMargin
                    anchors.rightMargin: searchLineEdit.anchors.leftMargin
                    anchors.top: parent.top
                    anchors.left: parent.left
                    anchors.bottom: parent.bottom
                    source: model.ico ? "image://" + model.ico : "image://empty"
                    sourceSize: Qt.size(width, height)
                    width: height
                }

                Text {
                    id: suggestText
                    anchors.top: parent.top
                    anchors.left: suggestIco.right
                    anchors.leftMargin: searchLineEdit.anchors.leftMargin
                    anchors.right: parent.right
                    anchors.bottom: parent.bottom
                    text: model.text
                    verticalAlignment: Text.AlignVCenter
                    font.pixelSize: searchLineEdit.font.pixelSize
                    wrapMode: Text.Wrap
                }

                MouseArea {
                    anchors.fill: parent
                    onClicked: {
                        model.candidate = true
                        searchLineEdit.acceptSearchCandidate()
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
        id: zoomLabel
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

    Rectangle {
        id: infoView
        width: parent.width
        y: parent.height
        height: previewHeight

        states: [
            State {
                name: "hidden"
            },
            State {
                name: "fullscreen"
            }
        ]

        ParallelAnimation {
            id: infoViewAnim
            property real targetY: 0
            property real targetHeight: 0
            running: false
            NumberAnimation {
                easing: Easing.InOutCubic
                duration: 200
                target: infoView
                property: "y"
                to: infoViewAnim.targetY
            }
            NumberAnimation {
                easing: Easing.InOutCubic
                duration: 200
                target: infoView
                property: "height"
                to: infoViewAnim.targetHeight
            }
        }

        property real previewHeight: parent.height * 0.2

        function setInfoData(d) {
            var text = cashpointTypePrintable(d.type)
            var bankJsonData = bankListModel.getBankData(d.bank)
            if (bankJsonData) {
                var bank = JSON.parse(bankJsonData)
                text += " " + bank.name
            }

            pointInfo.text = text

            var cashpointJsonData = cashpointModel.getCashpointById(d.id)
            console.log(cashpointJsonData)
            if (cashpointJsonData) {
                var cp = JSON.parse(cashpointJsonData)
                pointSchedule.text = cp.schedule
            }

            pointCurrencyType.text = currencyTypePrintable(cp)
        }

        function show() {
            infoViewAnim.targetY = parent.height - previewHeight
            infoViewAnim.targetHeight = previewHeight
            infoViewAnim.start()
            state = ""
        }

        function showFullscreen() {
            infoViewAnim.targetY = 0.0
            infoViewAnim.targetHeight = topView.height
            infoViewAnim.start()
            state = "fullscreen"
        }

        function hide() {
            infoViewAnim.targetY = topView.height
            infoViewAnim.targetHeight = previewHeight
            infoViewAnim.start()
            state = "hidden"
        }

        Column {
            anchors.fill: parent

            Label {
                id: pointInfo
            }

            Label {
                id: pointSchedule
            }

            Row {
                Label {
                    text: qsTr("Валюта:")
                }

                Label {
                    id: pointCurrencyType
                }
            }
        }
    }

//    RectangularGlow {
//        visible: infoView.state == "shown"
//        z: infoView.z - 1
//        anchors.fill: infoView
//        glowRadius: searchLineEditContainer.height / 5
//        spread: 0.3
//        color: "#11000055"
//        cornerRadius: glowRadius
//    }
}

