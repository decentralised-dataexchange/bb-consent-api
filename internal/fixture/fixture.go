package fixture

import (
	"io"
	"os"
	"strings"

	"github.com/bb-consent/api/internal/image"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/user"
)

const AssetsPath = "/opt/bb-consent/api/assets/"

func loadImageAndReturnBytes(imagePath string) ([]byte, error) {
	// Open the JPEG file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil

}

func saveImageToDb(data []byte) (string, error) {
	// Save image to db
	imageId, err := image.Add(data)
	if err != nil {
		return "", err
	}

	return imageId, nil
}

func loadCoverImageAssets(orgId string) error {
	coverImagePath := AssetsPath + "cover.jpeg"

	// Convert cover image to bytes
	coverImageBytes, err := loadImageAndReturnBytes(coverImagePath)
	if err != nil {
		return err
	}

	// Save cover image to db
	coverImageId, err := saveImageToDb(coverImageBytes)
	if err != nil {
		return err
	}

	// Update cover image for organisation
	_, err = org.UpdateCoverImage(orgId, coverImageId)
	if err != nil {
		return err
	}

	return nil
}

func loadLogoImageAssets(orgId string) error {
	logoImagePath := AssetsPath + "logo.jpeg"

	// Convert logo image to bytes
	logoImageBytes, err := loadImageAndReturnBytes(logoImagePath)
	if err != nil {
		return err
	}

	// Save logo image to db
	logoImageId, err := saveImageToDb(logoImageBytes)
	if err != nil {
		return err
	}

	// Update logo image for organisation
	_, err = org.UpdateLogoImage(orgId, logoImageId)
	if err != nil {
		return err
	}

	return nil
}

func LoadOrganisationAdminAvatarImageAssets(u user.User, hostUrl string) (user.User, error) {
	// Check if avatar image is present
	if len(strings.TrimSpace(u.ImageID)) != 0 {
		return u, nil
	}

	avatarImagePath := AssetsPath + "avatar.jpeg"

	// Convert avatar image to bytes
	avatarImageBytes, err := loadImageAndReturnBytes(avatarImagePath)
	if err != nil {
		return user.User{}, err
	}

	// Save avatar image to db
	avatarImageId, err := saveImageToDb(avatarImageBytes)
	if err != nil {
		return user.User{}, err
	}

	// Update avatar image for organisation
	u.ImageURL = "https://" + hostUrl + "/v2/onboard/admin/avatarimage"
	u.ImageID = avatarImageId
	u, err = user.Update(u.ID.Hex(), u)
	if err != nil {
		return user.User{}, err
	}

	return u, nil
}

func LoadImageAssetsForSingleTenantConfiguration() error {
	// Get first organisation
	o, err := org.GetFirstOrganization()
	if err != nil {
		return err
	}

	// Check if cover image is present
	if len(strings.TrimSpace(o.CoverImageID)) == 0 {
		err = loadCoverImageAssets(o.ID.Hex())
		if err != nil {
			return err
		}
	}

	// Check if logo image is present
	if len(strings.TrimSpace(o.LogoImageID)) == 0 {
		err = loadLogoImageAssets(o.ID.Hex())
		if err != nil {
			return err
		}
	}

	return nil
}
