#include "banklistsqlmodel.h"
#include "townlistsqlmodel.h"
#include "cashpointsqlmodel.h"
#include "serverapi.h"
#include "icoimageprovider.h"
#include "locationservice.h"

#include <QApplication>
#include <QQmlApplicationEngine>
#include <QQmlContext>
#include <QSqlDatabase>
#include <QSqlQuery>
#include <QSqlError>
#include <QFile>
#include <QDebug>
#include <QtCore/QSettings>
#include <QtCore/QStandardPaths>

QStringList getSqlQuery(const QString &queryFileName)
{
    QFile queryFile(queryFileName);
    if (!queryFile.open(QFile::ReadOnly | QFile::Text))
    {
        QTextStream(stderr) << "Cannot open " << queryFileName << " file" << endl;
        return QStringList();
    }
    QString queryStr = queryFile.readAll();
    return queryStr.split(";", QString::SkipEmptyParts);
}

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    app.setAttribute(Qt::AA_SynthesizeMouseForUnhandledTouchEvents, false);
    app.setAttribute(Qt::AA_SynthesizeTouchForUnhandledMouseEvents, false);
    app.setOrganizationName("Agnia");
    app.setApplicationName("CashPoints");


    const QString path =
#if QT_VERSION >= QT_VERSION_CHECK(5, 5, 0)
            QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation);
#else
            QStandardPaths::writableLocation(QStandardPaths::ConfigLocation);
#endif
    QSettings settings(path, QSettings::NativeFormat);

    // bank list db
    const QString banksConnName = "banks";
    QSqlDatabase db = QSqlDatabase::addDatabase("QSQLITE", banksConnName);
    db.setDatabaseName(":memory:");
    if (!db.open())
    {
        qFatal("Cannot create inmem database. Abort.");
        return 1;
    }

    db.transaction();
    db.exec("CREATE TABLE banks (id integer primary key, name text, licence integer, "
                                "name_tr text, town text, rating integer, "
                                "name_tr_alt text, tel text, ico_path text, mine integer)");

    db.exec("CREATE TABLE towns (id integer primary key, name text, name_tr text, "
                                "region_id integer, regional_center integer, mine integer)");

    db.exec("CREATE TABLE regions (id integer primary key, name text)");

    db.exec("CREATE TABLE cp (id integer primary key, type text, bank_id integer, "
                             "town_id integer, cord_lon real, cord_lat real, address text, "
                             "address_comment text, metro_name text, main_office integer, "
                             "without_weekend integer, round_the_clock integer, "
                             "works_as_shop integer, rub integer, usd integer, "
                             "eur integer, cash_in integer)");
    db.commit();

    ServerApi *api = new ServerApi("localhost", 8080);

    const QStringList icons = {
        ":/icon/star.svg",
        ":/icon/star_gray.svg",
        ":/icon/aim.svg",
        ":/icon/zoom_in.svg",
        ":/icon/zoom_out.svg",
        ":/icon/marker.svg"
    };
    IcoImageProvider *imageProvider = new IcoImageProvider;
    for (const QString &icoPath : icons) {
        QFile file(icoPath);
        if (file.open(QIODevice::ReadOnly)) {
            QString resName = icoPath.split('/').last();
            imageProvider->loadSvgImage(resName, file.readAll());
            qDebug() << icoPath << "loaded as" << resName;
        }
    }

    BankListSqlModel *bankListModel =
            new BankListSqlModel(banksConnName, api, imageProvider, &settings);

    TownListSqlModel *townListModel =
            new TownListSqlModel(banksConnName, api, imageProvider, &settings);

    CashPointSqlModel *cashpointModel =
            new CashPointSqlModel(banksConnName, api, imageProvider, &settings);

    LocationService *locationService = new LocationService(&app);

    QQmlApplicationEngine engine;

    engine.addImportPath("qrc:/ui");
    engine.addImageProvider(QLatin1String("ico"), imageProvider);
    engine.rootContext()->setContextProperty("bankListModel", bankListModel);
    engine.rootContext()->setContextProperty("townListModel", townListModel);
    engine.rootContext()->setContextProperty("cashpointModel", cashpointModel);
    engine.rootContext()->setContextProperty("serverApi", api);
    engine.rootContext()->setContextProperty("locationService", locationService);
    engine.load(QUrl(QStringLiteral("qrc:/ui/main.qml")));

    QObject *appWindow = nullptr;
    for (QObject *obj : engine.rootObjects()) {
        if (obj->objectName() == "appWindow") {
            appWindow = obj;
            break;
        }
    }
    Q_ASSERT(appWindow);
    QObject::connect(api, SIGNAL(pong(bool)), appWindow, SIGNAL(pong(bool)));

    /// update bank and town list after successfull ping
    QMetaObject::Connection connection = QObject::connect(api, &ServerApi::pong,
    [&](bool ok) {
        static int attempts = 0;
        attempts++;
        if (ok) {
            qDebug() << "Connected to server";
            QObject::disconnect(connection);
            bankListModel->updateFromServer();
            townListModel->updateFromServer();
        } else {
            if (attempts > 3) {
                return;
            }
            api->ping();
        }
    });
    api->ping();

    const int exitStatus = app.exec();

    delete cashpointModel;
    delete townListModel;
    delete bankListModel;

    delete api;

    db.close();

    return exitStatus;
}
