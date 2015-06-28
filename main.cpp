#include <QApplication>
#include <QQmlApplicationEngine>
#include <QQmlContext>
#include <QSqlDatabase>
#include <QSqlQuery>
#include <QSqlError>
#include <QFile>
#include <QDebug>
#include <QTableView>

#include "banklistsqlmodel.h"

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    app.setAttribute(Qt::AA_SynthesizeMouseForUnhandledTouchEvents, false);
    app.setAttribute(Qt::AA_SynthesizeTouchForUnhandledMouseEvents, false);

    QQmlApplicationEngine engine;

    engine.addImportPath("qrc:/");

    // bank list db
    const QString banksConnName = "banks";
    QSqlDatabase db = QSqlDatabase::addDatabase("QSQLITE", banksConnName);
    db.setDatabaseName(":memory:");
    if (!db.open())
    {
        qFatal("Cannot create inmem database. Abort.");
        return 1;
    }

    QFile banksQueryFile(":/bank.sql");
    if (!banksQueryFile.open(QFile::ReadOnly | QFile::Text))
    {
        qFatal("Cannot open bank.sql file");
        return 2;
    }
    QString queryStr = banksQueryFile.readAll();
    QStringList q_list = queryStr.split(";", QString::SkipEmptyParts);

    QSqlQuery q;
    db.transaction();
    foreach (QString qStr, q_list) {
        db.exec(qStr);
    }
    db.commit();

    q.exec("SELECT name, url, tel, tel_description FROM banks");
    while (q.next())
    {
        qDebug() << q.value(2).toString();
    }

    BankListSqlModel *bankListModel = new BankListSqlModel(banksConnName);

    engine.rootContext()->setContextProperty("bankListModel", bankListModel);
    engine.load(QUrl("qrc:/LeftMenu.qml"));
//    engine.load(QUrl(QStringLiteral("qrc:/BanksList.qml")));
//    engine.load(QUrl(QStringLiteral("qrc:/main.qml")));

    const int exitStatus = app.exec();

    delete bankListModel;

    db.close();

    return exitStatus;
}
