import QtQuick 2.4
import QtQuick.Controls 1.3

ApplicationWindow {
    visible: true
    width: 800
    height: 600

Rectangle {
    anchors.fill: parent

//    Rectangle {
//        color: "blue"
//        id: bankFilterEditContainer
//        anchors.top: parent.top
//        anchors.topMargin: 20
//        anchors.left: parent.left
//        anchors.leftMargin: 10
//        anchors.right: parent.right
//        anchors.rightMargin: 10
//        anchors.bottomMargin: 20
//        width: parent.width
//        height: bankFilterEdit.font.pixelSize * 1.5

        TextInput {
            id: bankFilterEdit
//            anchors.fill: parent
            anchors.top: parent.top
            anchors.topMargin: 20
            anchors.left: parent.left
            anchors.leftMargin: 10
            anchors.right: parent.right
            anchors.rightMargin: 10
            anchors.bottomMargin: 20
//                    width: parent.width

            font.pixelSize: parent.height / (15 * 1.5) > 24 ? parent.height / (15 * 1.5) : 24

            property bool isUserTextShowed: false
            property string placeHolderText: "Имя банка, сайт, номер тел..."
            property string userText: ""

            wrapMode: Text.WordWrap

            Component.onCompleted: {
                text = placeHolderText
                color = "lightgray"
            }

            onFocusChanged: {
                if (bankFilterEdit.focus) {
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
                if (displayText === "" || displayText === placeHolderText)
                {
                    bankListModel.setFilter("")
                } else {
                    bankListModel.setFilter(displayText)
                }
            }
        } // TextInput
//    }

    ListView {
        id: bankListView

//        anchors.top: bankFilterEditContainer.bottom
        anchors.top: bankFilterEdit.bottom
        anchors.topMargin: 20
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.bottom: parent.bottom

        model: bankListModel
        delegate: Rectangle {
            id: itemContatiner
//            color: model.index % 2 == 1 ? "white" : "lightgray"
            height: itemText.contentHeight + itemText.width * 0.04
            width: parent.width
            Rectangle {
                anchors.top: parent.top
                color: "lightgray"
                height: 1
                width: parent.width
            }

            Label {
                id: itemText

                anchors.verticalCenter: parent.verticalCenter

                anchors.left: parent.left
                anchors.right: parent.right
//                anchors.top: parent.top
                anchors.rightMargin: parent.width * 0.02
                anchors.leftMargin: parent.width * 0.02
//                anchors.topMargin: width * 0.1
//                anchors.fill: parent
                verticalAlignment: Text.AlignVCenter
                text: model.bank_name.replace(bankFilterEdit.displayText,
                                              '<b>' + bankFilterEdit.displayText +
                                              '</b>')
                textFormat: Text.StyledText
                wrapMode: Text.WordWrap
                font.pixelSize: bankListView.height / (15 * 1.5) > 24 ? bankListView.height / (15 * 1.5) : 24
//                font.bold: true

//                Behavior on scale {
//                    NumberAnimation { duration: 100 }
//                }
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
//                        itemText.scale = 0.98
                        itemContatiner.state = "clicked"
                    } else {
                        itemContatiner.state = ""
                    }
                }
            }
        }
    }
} // Rectangle
} // ApplicationWindow
