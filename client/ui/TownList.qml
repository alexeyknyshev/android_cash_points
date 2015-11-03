import QtQuick 2.0
import QtQuick.Controls 1.3
import QtGraphicalEffects 1.0

//ApplicationWindow {
//    visible: true
//    width: 800
//    height: 600
//    visibility: "FullScreen"

Rectangle {
    id: topRect
    anchors.fill: parent
    color: "#EDEDED"

    signal townSelected(int id, string name)

    onParentChanged: {
        townListModel.setFilter("")
    }

    Rectangle {
        id: townFilterEditContainer

        z: townListView.z + 1

        anchors.top: parent.top
        anchors.topMargin: anchors.leftMargin
        anchors.left: parent.left
        anchors.leftMargin: parent.width * 0.02
        anchors.right: parent.right
        anchors.rightMargin: anchors.leftMargin
        anchors.bottomMargin: anchors.leftMargin

        radius: height / 20
        height: Math.min(townFilterEdit.contentHeight * 2,
                         Math.max(parent.width, parent.height) * 0.15)

        TextInput {
            id: townFilterEdit

            z: townListView.z + 1

            maximumLength: 30
            anchors.margins: contentHeight * 0.5
            anchors.top: parent.top
            anchors.left: parent.left
            anchors.right: clearButton.left
            anchors.bottom: parent.bottom

            font.pixelSize: Math.max(topRect.height, topRect.width) * 0.03

            property bool isUserTextShowed: false
            property string placeHolderText: qsTr("Город, область / край / республика ...")
            property string userText: ""

            wrapMode: Text.Wrap

            Component.onCompleted: {
                text = placeHolderText
                color = "lightgray"
            }

            onFocusChanged: {
                if (townFilterEdit.focus) {
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

            onDisplayTextChanged: {
                if (displayText === "" || displayText === placeHolderText) {
                    townListModel.setFilter("")
                } else {
                    townListModel.setFilter(displayText)
                }
            }
        } // TextInput

        Rectangle {
            id: clearButton
            color: "transparent"
            anchors.right: parent.right
            anchors.top: parent.top
            anchors.topMargin: townFilterEdit.topMargin
            anchors.bottom: parent.bottom
            anchors.bottomMargin: townFilterEdit.bottomMargin
            width: height
            opacity: townFilterEdit.isUserTextShowed && townFilterEdit.displayText != "" ? 1.0 : 0.0

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
                    townFilterEdit.userText = ""
                    townFilterEdit.text = ""
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
        anchors.fill: townFilterEditContainer
        glowRadius: townFilterEditContainer.height / 10
        spread: 0.2
        color: "#11000055"
        cornerRadius: glowRadius
    }

    Rectangle {
        anchors.top: townFilterEditContainer.bottom
        anchors.topMargin: townFilterEditContainer.anchors.topMargin
        anchors.left: parent.left
        anchors.leftMargin: townFilterEditContainer.anchors.leftMargin
        anchors.right: parent.right
        anchors.rightMargin: townFilterEditContainer.anchors.rightMargin
        anchors.bottom: parent.bottom
        anchors.bottomMargin: townFilterEditContainer.anchors.bottomMargin

        RectangularGlow {
            anchors.fill: townListViewContainer
            glowRadius: townFilterEditContainer.height / 10
            spread: 0.2
            color: "#11000055"
            cornerRadius: glowRadius
        }

        ScrollView {
            id: townListViewContainer
            anchors.fill: parent
            verticalScrollBarPolicy: Qt.ScrollBarAlwaysOff

            ListView {
                id: townListView
                anchors.fill: parent
                snapMode: ListView.SnapToItem

                model: townListModel
                delegate: Rectangle {
                    id: itemContatiner
                    height: (itemText.contentHeight * (itemText.lineCount + 2) / itemText.lineCount)
                    width: parent.width
                    Rectangle {
                        anchors.top: parent.top
                        color: "lightgray"
                        height: 1
                        width: parent.width
                    }

                    Image {
                        id: itemImage
//                        source:
                        smooth: true
                        fillMode: Image.PreserveAspectFit
                        anchors.left: parent.left
                        anchors.leftMargin: townFilterEdit.anchors.leftMargin
                        anchors.top: parent.top
                        anchors.topMargin: townFilterEdit.anchors.topMargin
//                        anchors.bottom: parent.bottom
                        anchors.bottomMargin: townFilterEdit.anchors.bottomMargin
                        height: townListView.height * 0.1
                        width: height
                    }

                    Label {
                        id: itemText

                        anchors.verticalCenter: parent.verticalCenter

                        anchors.left: itemImage.right
                        anchors.right: parent.right
                        anchors.rightMargin: townFilterEdit.anchors.rightMargin
                        anchors.leftMargin: townFilterEdit.anchors.leftMargin

                        verticalAlignment: Text.AlignVCenter
                        text: model.town_name.replace(townFilterEdit.displayText,
                                                      '<b>' + townFilterEdit.displayText +
                                                      '</b>')
                        textFormat: Text.StyledText
                        wrapMode: Text.WordWrap
                        font.pixelSize: Math.max(topRect.height, topRect.width) * 0.03
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
                        onHoveredChanged: {
                            if (containsMouse) {
                                itemContatiner.state = "clicked"
                            } else {
                                itemContatiner.state = ""
                            }
                        }
                        onClicked: {
                            console.log("selected town: " + model.town_name + " (" + model.town_id + ")")
                            topRect.townSelected(model.town_id, model.town_name)
                        }
                    }
                }
            }
        } // Rectangle
    }
}
//} // ApplicationWindow
