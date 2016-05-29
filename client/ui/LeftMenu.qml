import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Window 2.2

Rectangle {
    property bool hidden: false

    states: State {
        name: "hidden"
        when: hidden
        PropertyChanges {
            target: topItem
            x: -width
        }
    }
    transitions: [
        Transition {
            from: ""
            to: "hidden"
            PropertyAnimation {
                duration: 500
                easing.type: Easing.InOutQuart
                properties: "x"
            }
            onRunningChanged: {
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
                properties: "x"
            }
            onRunningChanged: {
                if (running) {
                    visible = true
                }
            }
        }
    ]

    id: topItem

    property bool isAboutToHide: false

    signal action(var act)

    MultiPointTouchArea {
        anchors.fill: parent

        Image {
            id: userAccountView
            anchors.top: parent.top
            anchors.left: parent.left
            anchors.right: parent.right
            height: parent.height * 0.25

            Image {
                id: userAccountImage
                anchors.top: parent.top
                anchors.left: parent.left
                anchors.margins: parent.height * 0.15
                width: parent.height * 0.5
                height: width
                source: "image://ico/user.svg"
                sourceSize: Qt.size(width, height)
            }

            Rectangle {
                anchors.top: userAccountImage.bottom
                anchors.topMargin: parent.height * 0.15
                anchors.left: parent.left
                anchors.leftMargin: parent.height * 0.1
                anchors.bottom: parent.bottom
                anchors.right: parent.right
                anchors.rightMargin: parent.height * 0.1

                color: "#427FED"
                radius: height * 0.1

                scale: googleLoginMA.pressed ? 0.99 : 1

                Behavior on scale {
                    NumberAnimation {
                        duration: 100
                    }
                }

                Image {
                    id: googleIco
                    anchors.top: parent.top
                    anchors.left: parent.left
                    anchors.bottom: parent.bottom
                    anchors.topMargin: parent.height * 0.1
                    anchors.bottomMargin: parent.height * 0.1

                    width: height

                    sourceSize: Qt.size(width, height)
                    source: "image://ico/google.svg"
                }

                Rectangle {
                    id: separator
                    color: "white"
                    width: 1
                    anchors.left: googleIco.right
                    anchors.top: parent.top
                    anchors.bottom: parent.bottom
                }

                Label {
                    anchors.left: googleIco.right
                    anchors.right: parent.right
                    anchors.top: parent.top
                    anchors.bottom: parent.bottom

                    verticalAlignment: Qt.AlignVCenter
                    horizontalAlignment: Qt.AlignHCenter

                    text: qsTr("Вход Google")
                    color: "white"
                }

                MouseArea {
                    id: googleLoginMA
                    anchors.fill: parent

                    onClicked: {
                        googleApi.dial()
                    }
                }
            }
        }

/*
        Rectangle {
            id: userAccountView
            anchors.top: parent.top
            anchors.left: parent.left
            anchors.right: parent.right
            height: parent.height * 0.3

            Label {
                id: registerInfo
                anchors.top: parent.top
                anchors.left: parent.left
                anchors.right: parent.right
                anchors.margins: 20

                wrapMode: Text.WrapAtWordBoundaryOrAnywhere
                text: qsTr("<b>Зарегистрируйтесь!</b>\nРегистрация вам позволит...!")
            }

            Rectangle {
                id: registerButton
                anchors.top: registerInfo.bottom
                anchors.left: parent.left
                anchors.right: parent.right
    //            anchors.bottom: parent.bottom
                anchors.margins: 20
                visible: topItem.visible

                height: buttonText.contentHeight * 5

                color: "#3295BA"

                Text {
                    id: buttonText
                    anchors.fill: parent
                    horizontalAlignment: Text.AlignHCenter
                    verticalAlignment: Text.AlignVCenter
                    text: qsTr("Зарегистрироваться")
                    color: "white"
                    fontSizeMode: Text.VerticalFit
                    renderType: Text.NativeRendering
                    visible: topItem.visible
                }

                states: State {
                    name: "clicked"
                    PropertyChanges {
                        target: registerButton
                        color: "#29BCF2"
                    }
                }

                transitions: Transition {
                    ColorAnimation {
                        duration: 100
                    }
                }

                MouseArea {
                    anchors.fill: parent
                    onPressed: {
                        registerButton.state = "clicked"
                        console.log("pressed")
                    }
                    onReleased: {
                        registerButton.state = ""
                    }
                }
            }
        }
*/
        ListModel {
            id: menuModel

            ListElement {
                name: "bankSelectionItem"
                text: qsTr("Мои банки")
                ico: "../icon/bank.svg"
                qmlfile: "BanksList.qml"
            }
            ListElement {
                name: "townSelectionItem"
                text: qsTr("Города")
                ico: "../icon/town.svg"
                qmlfile: "TownList.qml"
            }
            ListElement {
                name: "settingsSelectionItem"
                text: qsTr("Настройки")
                ico: "../icon/settings.svg"
                qmlfile: ""
            }
            ListElement {
                name: "helpSelectonItem"
                text: qsTr("Помощь")
                ico: "../icon/info.svg"
                qmlfile: ""
            }
            ListElement {
                name: "feedbackSelectionItem"
                text: qsTr("Оставить отзыв")
                ico: "../icon/like.svg"
                qmlfile: ""
            }
            ListElement {
                name: "bugreportSelectionItem"
                text: qsTr("Сообщить об ошибке")
                ico: "../icon/bug.svg"
                qmlfile: ""
                url: "https://github.com/alexeyknyshev/android_cash_points/issues/new"
            }
        }

        ScrollView {
            anchors.top: userAccountView.bottom
            anchors.left:parent.left
            anchors.right: parent.right
            anchors.bottom: parent.bottom
            verticalScrollBarPolicy: Qt.ScrollBarAlwaysOff

            ListView {
                id: menu
                model: menuModel
                delegate: Rectangle {
                    z: parent.z + 1
                    id: itemContatiner
                    height: menu.height * (Screen.primaryOrientation == Screen.orientation ? 0.125 : 0.2)
                    width: parent.width

                    color: "white"

                    Image {
                        id: itemImageView
                        anchors.top: parent.top
                        anchors.topMargin: itemContatiner.height * 0.3
                        anchors.left: parent.left
                        anchors.rightMargin: anchors.topMargin * 2
                        anchors.leftMargin: anchors.topMargin
                        anchors.bottom: parent.bottom
                        anchors.bottomMargin: anchors.topMargin
                        //fillMode: Image.PreserveAspectFit

                        width: height
                        source: model.ico
                        sourceSize: Qt.size(width, height)
                    }

                    Label {
                        id: itemTextView
                        anchors.top: parent.top
                        anchors.right: parent.right
                        anchors.rightMargin: height * 0.25
                        anchors.left: itemImageView.right
                        anchors.leftMargin: height * 0.25
                        anchors.bottom: parent.bottom

                        text: model.text
                        font.weight: Font.DemiBold
                        verticalAlignment: Text.AlignVCenter
                        wrapMode: Text.WrapAtWordBoundaryOrAnywhere
                    }

                    states: State {
                        name: "clicked"
                        PropertyChanges {
                            target: itemContatiner
                            color: "lightgray"
                        }
                    }

                    transitions: Transition {
                        ColorAnimation {
                            duration: 100
                        }
                    }

                    MouseArea {
                        anchors.fill: parent
                        propagateComposedEvents: true
                        onHoveredChanged: {
                            if (containsMouse) {
                                itemContatiner.state = "clicked"
                            } else {
                                itemContatiner.state = ""
                            }
                        }
                        onClicked: {
                            if (model.qmlfile && model.qmlfile.length > 0) {
                                var act = {
                                    "type": "openView",
                                    "path": model.qmlfile
                                }
                                if (model.name === "townSelectionItem") {
                                    act["callback"] = function(t) {
                                        var townData = townListModel.getTownData(t.id)
                                        var town = JSON.parse(townData)
                                        if (town.id) {
                                            topItem.action({ "type": "undo" })
                                            topItem.action({
                                                               "type": "moveToCoord",
                                                               "coord": {
                                                                   "latitude": town.latitude,
                                                                   "longitude": town.longitude
                                                               },
                                                               "zoom": town.zoom
                                                           })
                                            topItem.state = "hidden"
                                        }
                                    }
                                }

                                topItem.action(act)
                                console.log(model.name + " clicked: loading " + model.qmlfile)
                            } else if (model.url && model.url.length > 0) {
                                feedbackService.openUrl(model.url)
                            }
                        }
                    }
                }
            }
        }
    }
}

//} // ApplicationWindow

