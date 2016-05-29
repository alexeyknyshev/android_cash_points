#include "banklistsqlmodel.h"
#include "townlistsqlmodel.h"
#include "cashpointsqlmodel.h"
#include "filtersmodel.h"
#include "serverapi.h"
#include "icoimageprovider.h"
#include "emptyimageprovider.h"
#include "locationservice.h"
#include "feedbackservice.h"
#include "searchengine.h"
#include "appstateproxy.h"
#include "hostsmodel.h"
#include "googleapiclient.h"

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

#include <QtCore/QJsonDocument>
#include <QtCore/QJsonObject>
#include <QtCore/QJsonArray>

class MyBanksDynamicFilter : public FiltersModel::DynamicFilter
{
public:
    MyBanksDynamicFilter(BankListSqlModel *model)
        : mBanksModel(model)
    {}

    QString createFilter()
    {
        auto banks = mBanksModel->getMineBanks();
        QJsonArray bank_id;
        for (int id : banks) {
            bank_id.append(id);
        }

        QJsonObject obj;
        obj["bank_id"] = bank_id;
        return QString::fromUtf8(QJsonDocument(obj).toJson());
    }

private:
    BankListSqlModel *mBanksModel;
};

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
    qDebug() << "Settings: " << settings.fileName();

    // bank list db
    const QString dbName = "banks";
    QSqlDatabase db = QSqlDatabase::addDatabase("QSQLITE", dbName);
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

    db.exec("CREATE TABLE partners (id integer, partner_id integer)");

    db.exec("CREATE TABLE towns (id integer primary key, name text, name_tr text, "
                                "region_id integer, regional_center integer, mine integer, "
                                "cord_lon real, cord_lat real, zoom real)");

    db.exec("CREATE TABLE regions (id integer primary key, name text)");

    db.exec("CREATE TABLE cp (id integer primary key, type text, bank_id integer, "
                             "town_id integer, cord_lon real, cord_lat real, address text, "
                             "address_comment text, metro_name text, main_office integer, "
                             "without_weekend integer, round_the_clock integer, "
                             "works_as_shop integer, free_access integer, currency text,"
                             "cash_in integer, timestamp integer, approved integer, "
                             "schedule text, patch_count integer)");
    db.commit();

    qRegisterMetaType<ServerApiPtr>("ServerApiPtr");
    ServerApi *api = new ServerApi(
//                                   "192.168.1.126"
//                                   "localhost"
//                                   "52.89.4.111"
                                 "5.23.98.144"
                                   , 8080);

    HostsModel *hostsModel = new HostsModel(api, &settings, api);

    const QStringList icons = {
        ":/icon/star.svg",
        ":/icon/star_gray.svg",
        ":/icon/aim.svg",
        ":/icon/zoom_in.svg",
        ":/icon/zoom_out.svg",
        ":/icon/marker.svg",
        ":/icon/place.svg",
        ":/icon/place_gray.svg",
        ":/icon/place_add.svg",
        ":/icon/place_add_plus.svg",
        ":/icon/add.svg",
        ":/icon/cluster.svg",
        ":/icon/round_the_clock.svg",
        ":/icon/limited_access.svg",
        ":/icon/editing.svg",
        ":/icon/eye.svg",
        ":/icon/clear.svg",
        ":/icon/share.svg",
        ":/icon/user.svg",
        ":/icon/google.svg"
    };
    const QMap<QString, QByteArray> iconsTemplates = {
        { ":/templates/event.svg", "#n" }
    };

    IcoImageProvider *icoImageProvider = new IcoImageProvider;
    for (const QString &icoPath : icons) {
        QFile file(icoPath);
        if (file.open(QIODevice::ReadOnly)) {
            QString resName = icoPath.split('/').last();
            icoImageProvider->loadSvgImage(resName, file.readAll());
            qDebug() << icoPath << "loaded as" << resName;
        } else {
            qDebug() << icoPath << "cannot load ico to image provider";
        }
    }
    for (auto it = iconsTemplates.cbegin(); it != iconsTemplates.cend(); it++) {
        const QString &icoPath = it.key();
        QFile file(icoPath);
        if (file.open(QIODevice::ReadOnly)) {
            QString resName = icoPath.split('/').last();
            icoImageProvider->loadSvgImageTemplate(resName, file.readAll(), it.value());
            qDebug() << icoPath << "template loaded as" << resName;
        } else {
            qDebug() << icoPath << "cannot load ico template to image provider";
        }
    }

    EmptyImageProvider *emptyImageProvider = new EmptyImageProvider;

    BankListSqlModel *bankListModel =
            new BankListSqlModel(dbName, api, icoImageProvider, &settings);

    TownListSqlModel *townListModel =
            new TownListSqlModel(dbName, api, icoImageProvider, &settings);

    CashPointSqlModel *cashpointModel =
            new CashPointSqlModel(dbName, api, icoImageProvider, &settings);

    FiltersModel *filtersModel =
            new FiltersModel(dbName, api, icoImageProvider, &settings);
    int id = filtersModel->addDynamicFilter(new MyBanksDynamicFilter(bankListModel));
    filtersModel->setFilterName(id, QObject::trUtf8("Мои банки"));

    SearchEngine *searchEngine = new SearchEngine(bankListModel, townListModel);

    LocationService *locationService = new LocationService(&app);
    FeedbackService *feedbackService = new FeedbackService(&app);
    GoogleApiClient *googleApiClient = new GoogleApiClient(&app);

    QQmlApplicationEngine *engine = new QQmlApplicationEngine;

    engine->addImportPath("qrc:/ui");
    engine->addImportPath("qrc:/");
    engine->addImageProvider(QLatin1String("ico"), icoImageProvider);
    engine->addImageProvider(QLatin1String("empty"), emptyImageProvider);
    engine->rootContext()->setContextProperty("bankListModel", bankListModel);
    engine->rootContext()->setContextProperty("townListModel", townListModel);
    engine->rootContext()->setContextProperty("cashpointModel", cashpointModel);
    engine->rootContext()->setContextProperty("serverApi", api);
    engine->rootContext()->setContextProperty("locationService", locationService);
    engine->rootContext()->setContextProperty("feedbackService", feedbackService);
    engine->rootContext()->setContextProperty("searchEngine", searchEngine);
    engine->rootContext()->setContextProperty("hostsModel", hostsModel);
    engine->rootContext()->setContextProperty("googleApi", googleApiClient);
    engine->rootContext()->setContextProperty("filtersModel", filtersModel);
    engine->load(QUrl(QStringLiteral("qrc:/ui/main.qml")));

    QObject *appWindow = nullptr;
    for (QObject *obj : engine->rootObjects()) {
        if (obj->objectName() == "appWindow") {
            appWindow = obj;
            break;
        }
    }
    Q_ASSERT(appWindow);
    QObject::connect(api, SIGNAL(pong(bool)), appWindow, SIGNAL(pong(bool)));

    AppStateProxy *proxy = new AppStateProxy(&app);
    QObject::connect(proxy, SIGNAL(appStateChanged(int)), appWindow, SIGNAL(appStateChanged(int)));
    QObject::connect(proxy, SIGNAL(serverDataLoaded(bool,QString)), appWindow, SIGNAL(serverDataReceived(bool,QString)));
    QObject::connect(bankListModel, SIGNAL(updateProgress(int,int)), appWindow, SIGNAL(banksUpdateProgress(int,int)));
    QObject::connect(townListModel, SIGNAL(updateProgress(int,int)), appWindow, SIGNAL(townsUpdateProgress(int,int)));

    /// update bank and town list after successfull ping
    QMetaObject::Connection connection = QObject::connect(api, &ServerApi::pong,
    [&](bool ok) {
        static int attempts = 0;
        attempts++;
        if (ok) {
            QObject::disconnect(connection);
            qDebug() << "Connected to server";

            QObject::connect(bankListModel, &BankListSqlModel::serverDataReceived,
                             proxy, &AppStateProxy::onBanksDataLoaded);
            QObject::connect(townListModel, &TownListSqlModel::serverDataReceived,
                             proxy, &AppStateProxy::onTownsDataLoaded);
            /*QObject::connect(bankListModel, &BankListSqlModel::requestError,
                             proxy, &AppStateProxy::onConnectionFailed);
            QObject::connect(townListModel, &TownListSqlModel::requestError,
                             proxy, &AppStateProxy::onConnectionFailed);*/
            bankListModel->updateFromServer();
            townListModel->updateFromServer();
        } else {
            if (attempts > 3) {
                proxy->onConnectionFailed();
                return;
            }
            api->ping();
        }
    });
    api->ping();

    const int exitStatus = app.exec();

    delete engine;

    delete searchEngine;
    delete cashpointModel;
    delete townListModel;
    delete bankListModel;

    delete api;

    db.close();

    return exitStatus;
}
