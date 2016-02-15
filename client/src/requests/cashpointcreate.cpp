#include "cashpointcreate.h"

#include <QtCore/QJsonObject>
#include <QtCore/QJsonArray>
#include <QtCore/QJsonParseError>
#include <QtCore/QDebug>

#include "../serverapi.h"

CashPointCreate::CashPointCreate(CashPointSqlModel *model)
    : CashPointRequest(model)
{
    registerStepHandlers({
                             STEP_HANDLER(createCashpoint)
                         });
}

#define CHECK_JSON_TYPE_STRING(val)\
if (!val.isString()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_NUMBER(val)\
if (!val.isDouble()) {\
    return false;\
}\

#define CHECK_JSON_TYPE_BOOL(val)\
if (!val.isBool()) {\
    return false;\
}\

bool CashPointCreate::fromJson(const QJsonObject &json)
{
    const QJsonValue type =   json["type"];
    const QJsonValue bankId = json["bank_id"];
    const QJsonValue townId = json["town_id"];
    const QJsonValue longitude = json["longitude"];
    const QJsonValue latitude = json["latitide"];
    const QJsonValue address = json["address"];
    const QJsonValue addressComment = json["address_comment"];
    const QJsonValue metroName = json["metro_name"];
    const QJsonValue freeAccess = json["free_access"];
    const QJsonValue mainOffice = json["main_office"];
    const QJsonValue withoutWeekend = json["without_weekend"];
    const QJsonValue roundTheClock = json["round_the_clock"];
    const QJsonValue worksAsShop = json["works_as_shop"];
    const QJsonValue schedule = json["schedule"];
    const QJsonValue tel = json["tel"];
    const QJsonValue additional = json["additional"];
    const QJsonValue rub = json["rub"];
    const QJsonValue usd = json["usd"];
    const QJsonValue eur = json["eur"];
    const QJsonValue cashIn = json["cash_in"];

    CHECK_JSON_TYPE_STRING(type)
    CHECK_JSON_TYPE_NUMBER(bankId)
    CHECK_JSON_TYPE_NUMBER(townId)
    CHECK_JSON_TYPE_NUMBER(longitude)
    CHECK_JSON_TYPE_NUMBER(latitude)
    CHECK_JSON_TYPE_STRING(address)
    CHECK_JSON_TYPE_STRING(addressComment)
    CHECK_JSON_TYPE_STRING(metroName)
    CHECK_JSON_TYPE_BOOL(freeAccess)
    CHECK_JSON_TYPE_BOOL(mainOffice)
    CHECK_JSON_TYPE_BOOL(withoutWeekend)
    CHECK_JSON_TYPE_BOOL(roundTheClock)
    CHECK_JSON_TYPE_BOOL(worksAsShop)
    CHECK_JSON_TYPE_STRING(schedule)
    CHECK_JSON_TYPE_STRING(tel)
    CHECK_JSON_TYPE_STRING(additional)
    CHECK_JSON_TYPE_BOOL(rub)
    CHECK_JSON_TYPE_BOOL(usd)
    CHECK_JSON_TYPE_BOOL(eur)
    CHECK_JSON_TYPE_BOOL(cashIn)

    return true;
}

void CashPointCreate::createCashpoint(ServerApiPtr api, quint32 leftAttempts)
{
    const int step = 0;

    api->sendRequest("/cashpoint", data,
    [this, api, leftAttempts](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (!isRunning()) {
            if (isDisposing()) {
                deleteLater();
            }
            return;
        }

        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        const QString errText = trUtf8("Cannot add new cashpoint to system");
        if (reqCode != ServerApi::RSC_Ok) {
            emitError(ServerApi::requestStatusCodeText(reqCode));
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitError("CashPointCreate: server response json parse error: " + err.errorString());
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (!json.isObject()) {
            emitError("CashPointCreate: response is not an object!");
            emitStepFinished(*api, step, false, errText);
            return;
        }

        const QJsonObject obj = json.object();
        if (!obj["cash_points"].isArray()) {
            emitError("CashPointCreate: cash_points field is not an array!");
            emitStepFinished(*api, step, false, errText);
            return;
        }

        mResponse = new CashPointResponse;

        const QJsonArray arr = obj["cash_points"].toArray();
        const auto end = arr.constEnd();
        for (auto it = arr.constBegin(); it != end; it++) {
            const QJsonValue &val = *it;
            const int id = val.toInt();
            if (id > 0) {
                mResponse->addCashPointData(this->data);
                mResponse->message = trUtf8("Cashpoint successfully added to system");
            }
        }
    });
}
