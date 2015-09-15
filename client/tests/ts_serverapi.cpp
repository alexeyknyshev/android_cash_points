#include <QtTest/QtTest>

#include <QtNetwork/QSslSocket>

#include "../src/serverapi.h"

class ServerApiTest : public QObject
{
    Q_OBJECT

private slots:
    void initTestCase();
    void cleanupTestCase();

    void testTownsRequest();
};

void ServerApiTest::initTestCase()
{
}

void ServerApiTest::cleanupTestCase()
{
}

void ServerApiTest::testTownsRequest()
{
//    QFile certFile(":/data/cert.pem");
//    QVERIFY2(certFile.open(QIODevice::ReadOnly), "Cannot open cert file");
//    QSslSocket::addDefaultCaCertificate(QSslCertificate(&certFile, QSsl::Pem));

    const int townId = 32;
    QVariantMap response;

    bool requestFinished = false;
    ServerApi api("127.0.0.1", 8080);
//    ServerApi api("127.0.0.1", 8081, &certFile);
    {
        QEventLoop loop;
        connect(&api, SIGNAL(requestTimedout(qint64)), &loop, SLOT(quit()));
        connect(&api, SIGNAL(responseReceived(qint64)), &loop, SLOT(quit()));

        QTimer *apiUpdateTimer = new QTimer(&api);
        connect(apiUpdateTimer, SIGNAL(timeout()), &api, SLOT(update()));

        api.setCallbacksExpireTime(1000);
        api.sendRequest("/town/" + QString::number(32), QJsonObject(), [&](const QByteArray &data, bool timeOut)
        {
            if (!timeOut) {
                requestFinished = true;
                QJsonDocument jsonDoc = QJsonDocument::fromJson(data);
                response = jsonDoc.object().toVariantMap();
            } else {
                qWarning() << "Request timed out";
                requestFinished = false;
            }
            loop.quit();
        });

        apiUpdateTimer->start(500);

        loop.exec();
    }

    QCOMPARE(requestFinished, true);
    {
        QCOMPARE(response["id"].toInt(),         townId);
        QCOMPARE(response["name"].toString(),    QString("Волоколамск"));
        QCOMPARE(response["name_tr"].toString(), QString("Volokolamsk"));
        QCOMPARE(response["zoom"].toInt(),       12);
    }
}

QTEST_MAIN(ServerApiTest)
#include "ts_serverapi.moc"
