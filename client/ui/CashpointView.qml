import QtQuick 2.0
import QtQuick.Controls 1.4

Rectangle {
    id: topRect

    property bool editingMode: false

    Column {
        anchors.fill: parent

        Label {
           text: qsTr("Id метки")
        }

        Label {
            id: idLabel
        }

        Label {
            text: qsTr("Тип")
        }

        TabView {
            tabsVisible: false
            currentIndex: editingMode ? 0 : 1
            Tab {
                Label {
                    text: typeComboBox.currentText
                }
            }
            Tab {
                ComboBox {
                    id: typeComboBox
                    model: [ qsTr("Банкомат"), qsTr("Офис"), qsTr("Касса") ]
                }
            }
        }

        Label {
            text: qsTr("Банк")
        }

        TabView {
            tabsVisible: false
            currentIndex: editingMode ? 0 : 1
            Tab {
                Label {
                    text: bankSelector.text
                }
            }
            Tab {
                Button {
                    id: bankSelector
                    property int bankId
                    onClicked: {

                    }
                }
            }
        }

        Label {
            text: qsTr("Адрес (комментарий)")
        }

        TextEdit {
            id: addressTextEdit
            readOnly: !editingMode
        }

        Row {
            Label {
                text: qsTr("В свободном доступе")
            }
            CheckBox {
                id: freeAccessCheckBox
            }
        }

        Row {
            Label {
                text: qsTr("Без выходных")
            }
            CheckBox {
                id: withoutWeekendCheckBox
                enabled: editingMode
            }
        }

        Row {
            Label {
                text: qsTr("Работает в режиме точки установки")
            }
            CheckBox {
                id: worksAsShopCheckBox
                enabled: editingMode
            }
        }

        Row {
            Label {
                text: qsTr("Рубли")
            }
            CheckBox {
                id: rubCheckBox
                enabled: editingMode
            }
        }

        Row {
            Label {
                text: qsTr("Доллары")
            }
            CheckBox {
                id: usdCheckBox
                enabled: editingMode
            }
        }

        Row {
            Label {
                text: qsTr("Евро")
            }
            CheckBox {
                id: eurCheckBox
                enabled: editingMode
            }
        }
    }
}

