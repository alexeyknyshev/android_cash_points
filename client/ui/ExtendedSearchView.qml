import QtQuick 2.0
import QtQuick.Layouts 1.1
import QtQuick.Controls 1.4
import QtQuick.Controls.Styles 1.4
import QtGraphicalEffects 1.0

import "currency.js" as Currency

Rectangle {
    id: topRect

    property var externalAction
    property var selectedBanks: []

    signal action(var act)

    onAction: {
        if (externalAction && externalAction.actionCallback) {
            externalAction.actionCallback(act)
        }
    }

    function generateColor(index, mix) {
        var frequency = .3;
//        for (var i = 0; i < 32; ++i)
//        {
           var r = Math.sin(frequency * index + 0) * 127 + 128;
           var g = Math.sin(frequency * index + 2) * 127 + 128;
           var b = Math.sin(frequency * index + 4) * 127 + 128;
//        }

        var r = r + mix.r / 2
        var g = g + mix.g / 2
        var b = b + mix.b / 2

        return Qt.rgba(r / 256, g / 256, b / 256, 0.5)
    }

    Flickable {
        id: flickable
        anchors.fill: parent
        anchors.margins: Math.min(topRect.width, topRect.height) * 0.03
        flickableDirection: Flickable.VerticalFlick
        contentWidth: grid.width
        contentHeight: grid.height

        Rectangle {
            id: grid
            height: searchButton.y + searchButton.height
            width: flickable.width

            Label {
                anchors.left: parent.left
                anchors.top: parent.top
                anchors.right: typeCombo.left
                height: typeCombo.height
                text: qsTr("Тип")
            }

            ComboBox {
                id: typeCombo
                anchors.right: parent.right
                anchors.top: parent.top
                height: flickable.height * 0.05
                width: flickable.width * 0.45
                model: ListModel {
                    id: typeModel
                    ListElement {
                        text: qsTr("Любой")
                    }
                    ListElement {
                        text: qsTr("Банкомат")
                        type: "atm"
                    }
                    ListElement {
                        text: qsTr("Офис")
                        type: "office"
                    }
                    ListElement {
                        text: qsTr("Касса")
                        type: "cash"
                    }
                }

                function currentType() {
                    return typeModel.get(find(currentText)).type
                }
            }

            Label {
                id: bankLabel
                anchors.top: typeCombo.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                anchors.right: partnersLabel.left
                height: flickable.height * 0.05

//                verticalAlignment: Text.AlignBottom
                horizontalAlignment: Text.AlignLeft

                text: qsTr("Банки")
            }

            Label {
                id: partnersLabel
                anchors.top: typeCombo.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: partnersCheck.left
                height: flickable.height * 0.05

//                verticalAlignment: Text.AlignBottom
                horizontalAlignment: Text.AlignRight

                text: qsTr("(и банки партнёры")
            }

            CheckBox {
                id: partnersCheck
                anchors.top: typeCombo.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: partnersLabelEnd.left
            }

            Label {
                id: partnersLabelEnd
                anchors.top: typeCombo.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                text: qsTr(")")

//                verticalAlignment: Text.AlignBottom
                horizontalAlignment: Text.AlignRight
            }

            Rectangle {
                id: bankViewRect
                anchors.top: bankLabel.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                anchors.right: parent.right
                height: flickable.height * 0.5

                RectangularGlow {
                    anchors.fill: bankListViewContainer
                    glowRadius: topRect.height  * 0.01
                    spread: 0.2
                    color: "#11000055"
                    cornerRadius: glowRadius
                }

                ScrollView {
                    id: bankListViewContainer
                    anchors.fill: parent
                    verticalScrollBarPolicy: Qt.ScrollBarAlwaysOff

                    onParentChanged: {
                        bankListModel.setFilter("", JSON.stringify({}))
                    }

                    onVisibleChanged: {
                        if (visible) {
                            bankListModel.setFilter("", JSON.stringify({}))
                        }
                    }

                    ListView {
                        id: bankListView
                        anchors.fill: parent
                        snapMode: ListView.SnapToItem

                        model: bankListModel
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
                                source: model.bank_ico_path ? "image://" + model.bank_ico_path : ""
                                sourceSize: Qt.size(width, height)
                                smooth: true
                                fillMode: Image.Pad
                                anchors.left: parent.left
                                anchors.leftMargin: parent.height * 0.2
                                anchors.top: parent.top
                                anchors.topMargin: parent.height * 0.2
                                height: bankListView.height * 0.1
                                width: height
                            }

                            Label {
                                id: itemText

                                anchors.verticalCenter: parent.verticalCenter

                                anchors.left: itemImage.right
                                anchors.right: parent.right
                                anchors.leftMargin: itemImage.width * 0.5

                                verticalAlignment: Text.AlignVCenter
                                text: model.bank_name/*.replace(bankFilterEdit.displayText,
                                                              '<b>' + bankFilterEdit.displayText +
                                                              '</b>')*/
                                textFormat: Text.StyledText
                                wrapMode: Text.WordWrap
                                font.pixelSize: Math.max(topRect.height, topRect.width) * 0.03
                            }

                            /*Image {
                                z: parent.z + 1
                                id: itemBankMine
                                source: model.bank_is_mine ? "image://ico/star.svg" : "image://ico/star_gray.svg"
                                sourceSize: Qt.size(width, height)
                                anchors.right: parent.right
                                anchors.top: parent.top
                                anchors.margins: Math.max(bankListView.width, bankListView.height) * 0.02
                                smooth: true
                                width: height
                                height: Math.max(bankListView.width, bankListView.height) * 0.06

                                MouseArea {
                                    z: parent.z + 1
                                    anchors.fill: parent
                                    onClicked: {
                                        model.bank_is_mine = model.bank_is_mine ? 0 : 1
                                    }
                                }
                            }*/

                            states: [
                                State {
                                    name: "clicked"
                                    PropertyChanges {
                                        target: itemContatiner
                                        color: "lightgray"
                                    }
                                },
                                State {
                                    name: "selected"
                                    PropertyChanges {
                                        target: itemContatiner
                                        color: "lightblue"
                                    }
                                    when: model.selected
                                }
                            ]

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
                                    } else if (itemContatiner.state == "clicked") {
                                        itemContatiner.state = ""
                                    }
                                }
                                onClicked: {
                                    console.log("selected bank: " + model.bank_name + " (" + model.bank_id + ") ")
                                    model.selected = !model.selected
                                }
                            }
                        }
                    }
                }
            }

        // TODO: banklist selector

        /*Label {
            text: qsTr("Населённый пункт")
        }

        // TODO: townSelector

        ComboBox {
            //model: townListModel
        }*/

//        Label {
//            text: qsTr("Время работы")
//        }

        // TODO: schedule selector

//        Label {
//            text: qsTr("Валюта")
//        }

//        ListView {
//            model: []
//        }

            Label {
                anchors.left: parent.left
                anchors.right: freeAccessCheck.right
                anchors.top: bankViewRect.bottom
                anchors.topMargin: flickable.anchors.margins
                height: flickable.height * 0.05

                text: qsTr("Только круглосуточные")
                wrapMode: Text.WrapAtWordBoundaryOrAnywhere
            }

            CheckBox {
                id: onlyRoundTheClockCheck
                anchors.top: bankViewRect.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                checked: false
            }

            Label {
                anchors.left: parent.left
                anchors.right: freeAccessCheck.right
                anchors.top: onlyRoundTheClockCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                height: flickable.height * 0.05

                text: qsTr("Только в свободном доступе")
                wrapMode: Text.WrapAtWordBoundaryOrAnywhere
            }

            CheckBox {
                id: freeAccessCheck
                anchors.top: onlyRoundTheClockCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                checked: false
            }

            Label {
                anchors.top: freeAccessCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                anchors.right: nearMetroCheck.left
                height: flickable.height * 0.05
                text: qsTr("Рядом с метро")
                wrapMode: Text.WrapAtWordBoundaryOrAnywhere
            }

            CheckBox {
                id: nearMetroCheck
                anchors.top: freeAccessCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                checked: false

                enabled: false
            }

            Label {
                anchors.top: nearMetroCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                anchors.right: onlyWithCashInCheck.left
                text: qsTr("Только с приёмом наличных")
                wrapMode: Text.WrapAtWordBoundaryOrAnywhere
            }

            CheckBox {
                id: onlyWithCashInCheck
                anchors.top: nearMetroCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                checked: false
            }

            Label {
                anchors.top: onlyWithCashInCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                text: qsTr("Рубли")
            }

            CheckBox {
                id: rubCheck
                anchors.top: onlyWithCashInCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                checked: false
            }

            Label {
                anchors.top: rubCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                text: qsTr("Доллары")
            }

            CheckBox {
                id: usdCheck
                anchors.top: rubCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                checked: false
            }

            Label {
                anchors.top: usdCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                text: qsTr("Евро")
            }

            CheckBox {
                id: eurCheck
                anchors.top: usdCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                checked: false
            }

            Label {
                anchors.top: eurCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                text: qsTr("Статус")
            }

            ComboBox {
                id: stateCombo
                anchors.top: eurCheck.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.right: parent.right
                height: flickable.height * 0.05
                width: flickable.width * 0.45
                model: ListModel {
                    id: stateModel
                    ListElement {
                        text: qsTr("Любой")
                        type: 0
                    }
                    ListElement {
                        text: qsTr("Утверждённые")
                        type: 1
                    }
                    ListElement {
                        text: qsTr("На рассмотрении")
                        type: 2
                    }
                }

                function currentType() {
                    return stateModel.get(find(currentText)).type
                }
            }

            Button {
                id: searchButton
                anchors.top: stateCombo.bottom
                anchors.topMargin: flickable.anchors.margins
                anchors.left: parent.left
                anchors.right: parent.right
                height: flickable.height * 0.075

                text: qsTr("найти")

                style: ButtonStyle {
                    background: Rectangle {
                        color: "#427FED"
                    }
                    label: Label {
                        color: "white"
//                        font.bold: true
                        font.pointSize: 24
                        text: control.text
                        verticalAlignment: Text.AlignVCenter
                        horizontalAlignment: Text.AlignHCenter
                    }
                }

                onClicked: {
                    var filter = {}

                    var banks = bankListModel.getSelectedIds()

                    if (banks.length > 0) {
                        var partners = []
                        if (partnersCheck.checked) {
                            partners = bankListModel.getPartnerBanks(banks)
                        }

                        var b = []
                        for (var i = 0; i < banks.length; i++) {
                            b.push(banks[i])
                        }
                        for (var j = 0; j < partners.length; j++) {
                            if (b.indexOf(partners[j]) == -1) {
                                b.push(partners[j])
                            }
                        }

                        filter["bank_id"] = b
                    }

                    if (onlyRoundTheClockCheck.checked) {
                        filter["round_the_clock"] = true
                    }

                    if (onlyWithCashInCheck.checked) {
                        filter["cash_in"] = true
                    }

                    var state = stateCombo.currentType()
                    switch (state) {
                    case 1: filter["approved"] = true; break
                    case 2: filter["approved"] = false; break;
                    }

                    if (freeAccessCheck.checked) {
                        filter["free_access"] = true
                    }

                    var type = typeCombo.currentType()
                    if (type) {
                        filter["type"] = type
                    }

                    var currency = []
                    if (rubCheck.checked) {
                        currency.push(Currency.RUB)
                    }

                    if (usdCheck.checked) {
                        currency.push(Currency.USD)
                    }

                    if (eurCheck.checked) {
                        currency.push(Currency.EUR)
                    }

                    topRect.action({
                                       "type": "filter",
                                       "saveAsRecent": true,
                                       "filter": filter
                                   })
                    topRect.action({ "type": "undo" })
                }
            }
        }
    }
}

