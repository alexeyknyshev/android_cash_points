#ifndef ICOIMAGEPROVIDER_H
#define ICOIMAGEPROVIDER_H

#include <QtQuick/QQuickImageProvider>

class QSvgRenderer;

class IcoImageProvider : public QQuickImageProvider
{
public:
    IcoImageProvider();
    ~IcoImageProvider();

    virtual QImage requestImage(const QString &id,
                                QSize *size,
                                const QSize &requestedSize) override;

    bool loadSvgImage(const QString &name, const QByteArray &data);
    void loadSvgImageTemplate(const QString &name, const QByteArray &data,
                              const QByteArray &placeholder);
    bool unloadSvgImage(const QString &name);

private:
    QMap<QString, QSvgRenderer *> mRenderers;
    QMap<QString, QPair<QByteArray, QByteArray>> mImageTemplates;
};

#endif // ICOIMAGEPROVIDER_H
