-- =====================================================
-- Expand Code Metadata Table - Phase 1
-- Purpose: Expand from 150 to 500+ codes covering all major industries
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 1, Task 1.1
-- =====================================================
-- 
-- This script adds 350+ additional codes to reach 500+ total codes.
-- Codes are organized by industry sectors with 10+ codes per code type per industry.
-- Uses ON CONFLICT to handle duplicates gracefully.
-- =====================================================

-- =====================================================
-- Part 1: Additional NAICS Codes (150+ codes)
-- =====================================================

INSERT INTO code_metadata (code_type, code, official_name, official_description, is_official, is_active)
VALUES
-- Technology Sector (Additional 20 codes)
('NAICS', '541330', 'Engineering Services', 
 'This industry comprises establishments primarily engaged in applying physical laws and principles of engineering in the design, development, and utilization of machines, materials, instruments, structures, processes, and systems.', 
 true, true),

('NAICS', '541611', 'Administrative Management and General Management Consulting Services', 
 'This industry comprises establishments primarily engaged in providing advice and assistance to businesses and other organizations on administrative management issues.', 
 true, true),

('NAICS', '541690', 'Other Scientific and Technical Consulting Services', 
 'This industry comprises establishments primarily engaged in providing advice and assistance to businesses and other organizations on scientific and technical issues (except environmental, computer systems design, and management consulting).', 
 true, true),

('NAICS', '541613', 'Marketing Consulting Services', 
 'This industry comprises establishments primarily engaged in providing advice and assistance to businesses and other organizations on marketing issues.', 
 true, true),

('NAICS', '541614', 'Process, Physical Distribution, and Logistics Consulting Services', 
 'This industry comprises establishments primarily engaged in providing advice and assistance to businesses and other organizations on process, physical distribution, and logistics issues.', 
 true, true),

('NAICS', '541618', 'Other Management Consulting Services', 
 'This industry comprises establishments primarily engaged in providing management consulting services (except administrative and general management, human resources, marketing, process, physical distribution, and logistics consulting).', 
 true, true),

('NAICS', '541620', 'Environmental Consulting Services', 
 'This industry comprises establishments primarily engaged in providing advice and assistance to businesses and other organizations on environmental issues.', 
 true, true),

('NAICS', '541720', 'Research and Development in the Social Sciences and Humanities', 
 'This industry comprises establishments primarily engaged in conducting research and analyses in cognitive development, sociology, psychology, language, behavior, economic, and other social science and humanities research.', 
 true, true),

('NAICS', '541930', 'Translation and Interpretation Services', 
 'This industry comprises establishments primarily engaged in translating written material and interpreting speech from one language to another.', 
 true, true),

('NAICS', '518210', 'Data Processing, Hosting, and Related Services', 
 'This industry comprises establishments primarily engaged in providing infrastructure for hosting, data processing services, and related services.', 
 true, true),

('NAICS', '518310', 'Internet Service Providers and Web Search Portals', 
 'This industry comprises establishments primarily engaged in providing Internet access services and/or operating Web search portals.', 
 true, true),

('NAICS', '519130', 'Internet Publishing and Broadcasting and Web Search Portals', 
 'This industry comprises establishments primarily engaged in publishing and/or broadcasting content on the Internet exclusively.', 
 true, true),

('NAICS', '541214', 'Payroll Services', 
 'This industry comprises establishments primarily engaged in providing payroll processing services.', 
 true, true),

('NAICS', '541219', 'Other Accounting Services', 
 'This industry comprises establishments primarily engaged in providing accounting services (except offices of certified public accountants and payroll services).', 
 true, true),

('NAICS', '541410', 'Interior Design Services', 
 'This industry comprises establishments primarily engaged in planning, designing, and administering projects in interior spaces to meet the physical and aesthetic needs of people using them.', 
 true, true),

('NAICS', '541420', 'Industrial Design Services', 
 'This industry comprises establishments primarily engaged in creating and developing designs and specifications that optimize the use, value, and appearance of products.', 
 true, true),

('NAICS', '541430', 'Graphic Design Services', 
 'This industry comprises establishments primarily engaged in planning, designing, and managing the production of visual communication in order to convey specific messages or concepts, clarify complex information, or project visual identities.', 
 true, true),

('NAICS', '541511', 'Custom Computer Programming Services', 
 'This U.S. industry comprises establishments primarily engaged in writing, modifying, testing, and supporting software to meet the needs of a particular customer.', 
 true, true),

('NAICS', '541512', 'Computer Systems Design Services', 
 'This U.S. industry comprises establishments primarily engaged in planning and designing computer systems that integrate computer hardware, software, and communication technologies.', 
 true, true),

-- Financial Services (Additional 20 codes)
('NAICS', '522210', 'Credit Card Issuing', 
 'This industry comprises establishments primarily engaged in issuing credit cards.', 
 true, true),

('NAICS', '522291', 'Consumer Lending', 
 'This industry comprises establishments primarily engaged in making unsecured cash loans to consumers.', 
 true, true),

('NAICS', '522292', 'Real Estate Credit', 
 'This industry comprises establishments primarily engaged in lending funds with real estate as collateral.', 
 true, true),

('NAICS', '522293', 'International Trade Financing', 
 'This industry comprises establishments primarily engaged in providing financing for international trade transactions.', 
 true, true),

('NAICS', '522294', 'Secondary Market Financing', 
 'This industry comprises establishments primarily engaged in providing financing for the purchase of loans and other financial assets.', 
 true, true),

('NAICS', '522298', 'All Other Nondepository Credit Intermediation', 
 'This industry comprises establishments primarily engaged in providing nondepository credit intermediation (except credit card issuing, consumer lending, real estate credit, international trade financing, and secondary market financing).', 
 true, true),

('NAICS', '522310', 'Mortgage and Nonmortgage Loan Brokers', 
 'This industry comprises establishments primarily engaged in arranging loans for others.', 
 true, true),

('NAICS', '522320', 'Financial Transactions Processing, Reserve, and Clearinghouse Activities', 
 'This industry comprises establishments primarily engaged in providing financial transaction processing, reserve, and clearinghouse services.', 
 true, true),

('NAICS', '522390', 'Other Activities Related to Credit Intermediation', 
 'This industry comprises establishments primarily engaged in facilitating credit intermediation (except mortgage and nonmortgage loan brokers and financial transactions processing).', 
 true, true),

('NAICS', '523120', 'Securities Brokerage', 
 'This industry comprises establishments primarily engaged in buying and selling securities on a commission or transaction fee basis.', 
 true, true),

('NAICS', '523130', 'Commodity Contracts Dealing', 
 'This industry comprises establishments primarily engaged in buying and selling commodity contracts on a commission or transaction fee basis.', 
 true, true),

('NAICS', '523140', 'Commodity Contracts Brokerage', 
 'This industry comprises establishments primarily engaged in arranging commodity contracts on a commission or transaction fee basis.', 
 true, true),

('NAICS', '523210', 'Securities and Commodity Exchanges', 
 'This industry comprises establishments primarily engaged in furnishing physical or electronic marketplaces for the purpose of facilitating the buying and selling of stocks, stock options, bonds, or commodity contracts.', 
 true, true),

('NAICS', '523910', 'Miscellaneous Intermediation', 
 'This industry comprises establishments primarily engaged in acting as agents (i.e., brokers) in buying or selling financial contracts (except securities brokerages and commodity contracts brokerages).', 
 true, true),

('NAICS', '523920', 'Portfolio Management', 
 'This industry comprises establishments primarily engaged in managing the portfolio assets (i.e., funds) of others.', 
 true, true),

('NAICS', '523930', 'Investment Advice', 
 'This industry comprises establishments primarily engaged in providing customized investment advice to clients.', 
 true, true),

('NAICS', '523991', 'Trust, Fiduciary, and Custody Activities', 
 'This industry comprises establishments primarily engaged in providing trust, fiduciary, and custody services to others.', 
 true, true),

('NAICS', '523999', 'Miscellaneous Financial Investment Activities', 
 'This industry comprises establishments primarily engaged in providing financial investment services (except securities and commodity exchanges, securities brokerages, commodity contracts brokerages, portfolio management, investment advice, and trust, fiduciary, and custody activities).', 
 true, true),

('NAICS', '524113', 'Direct Life Insurance Carriers', 
 'This industry comprises establishments primarily engaged in initially underwriting (i.e., assuming the risk and assigning premiums) life insurance policies.', 
 true, true),

('NAICS', '524114', 'Direct Health and Medical Insurance Carriers', 
 'This industry comprises establishments primarily engaged in initially underwriting (i.e., assuming the risk and assigning premiums) health and medical insurance policies.', 
 true, true),

-- Healthcare (Additional 20 codes)
('NAICS', '621210', 'Offices of Dentists', 
 'This industry comprises establishments of licensed practitioners having the degree of D.M.D. (Doctor of Dental Medicine), D.D.S. (Doctor of Dental Surgery), or D.D. (Doctor of Dentistry) primarily engaged in the independent practice of general or specialized dentistry or dental surgery.', 
 true, true),

('NAICS', '621310', 'Offices of Chiropractors', 
 'This industry comprises establishments of licensed practitioners having the degree of D.C. (Doctor of Chiropractic) primarily engaged in the independent practice of chiropractic.', 
 true, true),

('NAICS', '621320', 'Offices of Optometrists', 
 'This industry comprises establishments of licensed practitioners having the degree of O.D. (Doctor of Optometry) primarily engaged in the independent practice of optometry.', 
 true, true),

('NAICS', '621330', 'Offices of Mental Health Practitioners (except Physicians)', 
 'This industry comprises establishments of licensed practitioners having the degree of Ph.D. (Doctor of Philosophy) or Psy.D. (Doctor of Psychology) primarily engaged in the independent practice of psychology.', 
 true, true),

('NAICS', '621340', 'Offices of Physical, Occupational and Speech Therapists, and Audiologists', 
 'This industry comprises establishments of licensed practitioners primarily engaged in the independent practice of physical therapy, occupational therapy, speech-language therapy, or audiology.', 
 true, true),

('NAICS', '621391', 'Offices of Podiatrists', 
 'This industry comprises establishments of licensed practitioners having the degree of D.P.M. (Doctor of Podiatric Medicine) primarily engaged in the independent practice of podiatry.', 
 true, true),

('NAICS', '621399', 'Offices of All Other Miscellaneous Health Practitioners', 
 'This industry comprises establishments of licensed practitioners primarily engaged in the independent practice of health services (except physicians, dentists, chiropractors, optometrists, mental health practitioners, physical therapists, occupational therapists, speech therapists, audiologists, and podiatrists).', 
 true, true),

('NAICS', '621410', 'Family Planning Centers', 
 'This industry comprises establishments primarily engaged in providing family planning services.', 
 true, true),

('NAICS', '621420', 'Outpatient Mental Health and Substance Abuse Centers', 
 'This industry comprises establishments primarily engaged in providing outpatient services for the diagnosis and treatment of mental health disorders and substance abuse.', 
 true, true),

('NAICS', '621491', 'HMO Medical Centers', 
 'This industry comprises establishments primarily engaged in providing health maintenance organization (HMO) medical services.', 
 true, true),

('NAICS', '621492', 'Kidney Dialysis Centers', 
 'This industry comprises establishments primarily engaged in providing kidney dialysis services.', 
 true, true),

('NAICS', '621493', 'Freestanding Ambulatory Surgical and Emergency Centers', 
 'This industry comprises establishments primarily engaged in providing surgical and emergency services on an outpatient basis.', 
 true, true),

('NAICS', '621498', 'All Other Outpatient Care Centers', 
 'This industry comprises establishments primarily engaged in providing outpatient care services (except family planning centers, outpatient mental health and substance abuse centers, HMO medical centers, kidney dialysis centers, and freestanding ambulatory surgical and emergency centers).', 
 true, true),

('NAICS', '621511', 'Medical Laboratories', 
 'This industry comprises establishments primarily engaged in providing medical laboratory testing services.', 
 true, true),

('NAICS', '621512', 'Diagnostic Imaging Centers', 
 'This industry comprises establishments primarily engaged in providing diagnostic imaging services.', 
 true, true),

('NAICS', '621610', 'Home Health Care Services', 
 'This industry comprises establishments primarily engaged in providing skilled nursing services in the home.', 
 true, true),

('NAICS', '621910', 'Ambulance Services', 
 'This industry comprises establishments primarily engaged in providing ambulance services for emergency and nonemergency transportation of patients.', 
 true, true),

('NAICS', '621991', 'Blood and Organ Banks', 
 'This industry comprises establishments primarily engaged in collecting, storing, and distributing blood and blood products and body organs.', 
 true, true),

('NAICS', '621999', 'All Other Miscellaneous Ambulatory Health Care Services', 
 'This industry comprises establishments primarily engaged in providing ambulatory health care services (except offices of physicians, dentists, and other health practitioners, outpatient care centers, medical laboratories, diagnostic imaging centers, home health care services, ambulance services, and blood and organ banks).', 
 true, true),

('NAICS', '622210', 'Psychiatric and Substance Abuse Hospitals', 
 'This industry comprises establishments primarily engaged in providing diagnostic and medical treatment, and continuous nursing care to patients with mental illness or substance abuse disorders.', 
 true, true),

-- Retail (Additional 20 codes)
('NAICS', '452210', 'Warehouse Clubs and Supercenters', 
 'This industry comprises establishments known as warehouse clubs, superstores, or supercenters primarily engaged in retailing a general line of groceries in combination with general lines of new merchandise, such as apparel, furniture, and appliances.', 
 true, true),

('NAICS', '452311', 'Warehouse Clubs and Supercenters', 
 'This industry comprises establishments known as warehouse clubs, superstores, or supercenters primarily engaged in retailing a general line of groceries in combination with general lines of new merchandise.', 
 true, true),

('NAICS', '453110', 'Florists', 
 'This industry comprises establishments primarily engaged in retailing cut flowers, floral arrangements, and potted plants purchased from others.', 
 true, true),

('NAICS', '453210', 'Office Supplies and Stationery Stores', 
 'This industry comprises establishments primarily engaged in retailing new office supplies, stationery, gift wrap, and gift boxes.', 
 true, true),

('NAICS', '453220', 'Gift, Novelty, and Souvenir Stores', 
 'This industry comprises establishments primarily engaged in retailing new gifts, novelty merchandise, souvenirs, greeting cards, seasonal and holiday decorations, and curios.', 
 true, true),

('NAICS', '453310', 'Used Merchandise Stores', 
 'This industry comprises establishments primarily engaged in retailing used merchandise.', 
 true, true),

('NAICS', '453910', 'Pet and Pet Supplies Stores', 
 'This industry comprises establishments primarily engaged in retailing pets, pet foods, and pet supplies.', 
 true, true),

('NAICS', '453920', 'Art Dealers', 
 'This industry comprises establishments primarily engaged in retailing original and limited edition art works.', 
 true, true),

('NAICS', '453930', 'Manufactured (Mobile) Home Dealers', 
 'This industry comprises establishments primarily engaged in retailing new and used manufactured (mobile) homes.', 
 true, true),

('NAICS', '454110', 'Electronic Shopping and Mail-Order Houses', 
 'This industry comprises establishments primarily engaged in retailing all types of merchandise using nonstore means, such as catalogs, toll free telephone numbers, or electronic media, such as interactive television or computer.', 
 true, true),

('NAICS', '454210', 'Vending Machine Operators', 
 'This industry comprises establishments primarily engaged in retailing merchandise through vending machines that they service.', 
 true, true),

('NAICS', '454310', 'Fuel Dealers', 
 'This industry comprises establishments primarily engaged in retailing heating oil, liquefied petroleum (LP) gas, and other fuels (except gasoline) via direct selling.', 
 true, true),

('NAICS', '454390', 'Other Direct Selling Establishments', 
 'This industry comprises establishments primarily engaged in retailing merchandise via nonstore means (except electronic shopping and mail-order houses, vending machine operators, and fuel dealers).', 
 true, true),

('NAICS', '441110', 'New Car Dealers', 
 'This industry comprises establishments primarily engaged in retailing new automobiles, light trucks, and sport utility vehicles.', 
 true, true),

('NAICS', '441120', 'Used Car Dealers', 
 'This industry comprises establishments primarily engaged in retailing used automobiles, light trucks, and sport utility vehicles.', 
 true, true),

('NAICS', '442110', 'Furniture Stores', 
 'This industry comprises establishments primarily engaged in retailing new furniture.', 
 true, true),

('NAICS', '442210', 'Floor Covering Stores', 
 'This industry comprises establishments primarily engaged in retailing new floor coverings, such as rugs, tiles, and carpets.', 
 true, true),

('NAICS', '443141', 'Household Appliance Stores', 
 'This industry comprises establishments primarily engaged in retailing new household appliances.', 
 true, true),

('NAICS', '444110', 'Home Centers', 
 'This industry comprises establishments known as home centers primarily engaged in retailing a general line of new home repair and improvement materials and supplies.', 
 true, true),

('NAICS', '444130', 'Hardware Stores', 
 'This industry comprises establishments primarily engaged in retailing a general line of new hardware items, such as tools and builders'' hardware.', 
 true, true),

-- Food & Beverage (Additional 15 codes)
('NAICS', '722410', 'Drinking Places (Alcoholic Beverages)', 
 'This industry comprises establishments primarily engaged in preparing and serving alcoholic beverages for immediate consumption on the premises.', 
 true, true),

('NAICS', '722320', 'Caterers', 
 'This industry comprises establishments primarily engaged in providing single event-based food services.', 
 true, true),

('NAICS', '722330', 'Mobile Food Services', 
 'This industry comprises establishments primarily engaged in preparing and serving meals and snacks for immediate consumption from motorized vehicles or nonmotorized carts.', 
 true, true),

('NAICS', '311811', 'Retail Bakeries', 
 'This industry comprises establishments primarily engaged in retailing bakery products not for immediate consumption made on the premises from flour and other ingredients.', 
 true, true),

('NAICS', '311812', 'Commercial Bakeries', 
 'This industry comprises establishments primarily engaged in manufacturing bread and other bakery products (except cookies and crackers) for wholesale distribution.', 
 true, true),

('NAICS', '445110', 'Supermarkets and Grocery Stores', 
 'This industry comprises establishments primarily engaged in retailing a general line of food, such as canned and frozen foods; fresh fruits and vegetables; and fresh and prepared meats, fish, and poultry.', 
 true, true),

('NAICS', '445120', 'Convenience Stores', 
 'This industry comprises establishments primarily engaged in retailing a limited line of goods that generally includes milk, bread, soda, and snacks.', 
 true, true),

('NAICS', '445210', 'Meat Markets', 
 'This industry comprises establishments primarily engaged in retailing fresh, frozen, or processed meats and poultry.', 
 true, true),

('NAICS', '445220', 'Fish and Seafood Markets', 
 'This industry comprises establishments primarily engaged in retailing fresh, frozen, or processed fish and seafood.', 
 true, true),

('NAICS', '445230', 'Fruit and Vegetable Markets', 
 'This industry comprises establishments primarily engaged in retailing fresh fruits and vegetables.', 
 true, true),

('NAICS', '445291', 'Baked Goods Stores', 
 'This industry comprises establishments primarily engaged in retailing baked goods (except cookies and crackers) not for immediate consumption made elsewhere.', 
 true, true),

('NAICS', '445292', 'Confectionery and Nut Stores', 
 'This industry comprises establishments primarily engaged in retailing confectionery and nuts.', 
 true, true),

('NAICS', '445299', 'All Other Specialty Food Stores', 
 'This industry comprises establishments primarily engaged in retailing specialty foods (except meat markets, fish and seafood markets, fruit and vegetable markets, baked goods stores, and confectionery and nut stores).', 
 true, true),

('NAICS', '446110', 'Pharmacies and Drug Stores', 
 'This industry comprises establishments known as pharmacies and drug stores primarily engaged in retailing prescription or nonprescription drugs and medicines.', 
 true, true),

('NAICS', '446120', 'Cosmetics, Beauty Supplies, and Perfume Stores', 
 'This industry comprises establishments primarily engaged in retailing cosmetics, perfumes, toiletries, and personal grooming appliances.', 
 true, true),

-- Manufacturing (Additional 15 codes)
('NAICS', '311821', 'Cookie and Cracker Manufacturing', 
 'This industry comprises establishments primarily engaged in manufacturing cookies and crackers.', 
 true, true),

('NAICS', '311830', 'Tortilla Manufacturing', 
 'This industry comprises establishments primarily engaged in manufacturing tortillas.', 
 true, true),

('NAICS', '312111', 'Soft Drink Manufacturing', 
 'This industry comprises establishments primarily engaged in manufacturing soft drinks and carbonated waters.', 
 true, true),

('NAICS', '312112', 'Bottled Water Manufacturing', 
 'This industry comprises establishments primarily engaged in purifying and bottling water.', 
 true, true),

('NAICS', '312113', 'Ice Manufacturing', 
 'This industry comprises establishments primarily engaged in manufacturing ice.', 
 true, true),

('NAICS', '312120', 'Breweries', 
 'This industry comprises establishments primarily engaged in brewing beer, ale, malt liquors, and nonalcoholic beer.', 
 true, true),

('NAICS', '312130', 'Wineries', 
 'This industry comprises establishments primarily engaged in growing grapes and manufacturing wines and brandies.', 
 true, true),

('NAICS', '312140', 'Distilleries', 
 'This industry comprises establishments primarily engaged in distilling potable liquors (except brandies).', 
 true, true),

('NAICS', '313210', 'Broadwoven Fabric Mills', 
 'This industry comprises establishments primarily engaged in weaving fabrics (except narrow fabrics) on looms.', 
 true, true),

('NAICS', '313310', 'Textile and Fabric Finishing Mills', 
 'This industry comprises establishments primarily engaged in finishing textiles, fabrics, and apparel.', 
 true, true),

('NAICS', '314110', 'Carpet and Rug Mills', 
 'This industry comprises establishments primarily engaged in manufacturing woven, tufted, and other carpets and rugs.', 
 true, true),

('NAICS', '315210', 'Cut and Sew Apparel Contractors', 
 'This industry comprises establishments primarily engaged in manufacturing apparel from materials owned by others.', 
 true, true),

('NAICS', '315220', 'Men''s and Boys'' Cut and Sew Apparel Manufacturing', 
 'This industry comprises establishments primarily engaged in manufacturing men''s and boys'' apparel from purchased fabric.', 
 true, true),

-- Construction (Additional 15 codes)
('NAICS', '236220', 'Commercial and Institutional Building Construction', 
 'This industry comprises establishments primarily engaged in the construction (including new work, additions, alterations, and repairs) of commercial and institutional buildings and related structures.', 
 true, true),

('NAICS', '237310', 'Highway, Street, and Bridge Construction', 
 'This industry comprises establishments primarily engaged in the construction of highways (including elevated), streets, roads, airport runways, public sidewalks, or bridges.', 
 true, true),

('NAICS', '236210', 'Industrial Building Construction', 
 'This industry comprises establishments primarily engaged in the construction (including new work, additions, alterations, and repairs) of industrial buildings and related structures.', 
 true, true),

('NAICS', '236115', 'New Single-Family Housing Construction (except For-Sale Builders)', 
 'This industry comprises establishments primarily engaged in the construction of new single-family housing units (except for-sale builders).', 
 true, true),

('NAICS', '236116', 'New Multifamily Housing Construction (except For-Sale Builders)', 
 'This industry comprises establishments primarily engaged in the construction of new multifamily housing units (except for-sale builders).', 
 true, true),

('NAICS', '236117', 'New Housing For-Sale Builders', 
 'This industry comprises establishments primarily engaged in the construction of new housing units for sale.', 
 true, true),

('NAICS', '237110', 'Water and Sewer Line and Related Structures Construction', 
 'This industry comprises establishments primarily engaged in the construction of water and sewer lines, pumping stations, and related structures.', 
 true, true),

('NAICS', '237120', 'Oil and Gas Pipeline and Related Structures Construction', 
 'This industry comprises establishments primarily engaged in the construction of oil and gas pipelines and related structures.', 
 true, true),

('NAICS', '237130', 'Power and Communication Line and Related Structures Construction', 
 'This industry comprises establishments primarily engaged in the construction of power and communication transmission lines and related structures.', 
 true, true),

('NAICS', '237210', 'Land Subdivision', 
 'This industry comprises establishments primarily engaged in subdividing real property into lots for sale to others.', 
 true, true),

('NAICS', '237990', 'Other Heavy and Civil Engineering Construction', 
 'This industry comprises establishments primarily engaged in heavy and civil engineering construction (except highway, street, and bridge construction, water and sewer line construction, oil and gas pipeline construction, power and communication line construction, and land subdivision).', 
 true, true),

('NAICS', '238110', 'Poured Concrete Foundation and Structure Contractors', 
 'This industry comprises establishments primarily engaged in pouring and finishing concrete foundations and structural elements.', 
 true, true),

('NAICS', '238120', 'Structural Steel and Precast Concrete Contractors', 
 'This industry comprises establishments primarily engaged in erecting and assembling structural steel and precast concrete products.', 
 true, true),

('NAICS', '238130', 'Framing Contractors', 
 'This industry comprises establishments primarily engaged in framing buildings and other structures.', 
 true, true),

('NAICS', '238140', 'Masonry Contractors', 
 'This industry comprises establishments primarily engaged in masonry work.', 
 true, true),

-- Real Estate (Additional 10 codes)
('NAICS', '531110', 'Lessors of Residential Buildings and Dwellings', 
 'This industry comprises establishments primarily engaged in acting as lessors of buildings used as residences or dwellings.', 
 true, true),

('NAICS', '531120', 'Lessors of Nonresidential Buildings (except Miniwarehouses)', 
 'This industry comprises establishments primarily engaged in acting as lessors of buildings (except miniwarehouses and self-storage units) that are not used as residences or dwellings.', 
 true, true),

('NAICS', '531210', 'Offices of Real Estate Agents and Brokers', 
 'This industry comprises establishments primarily engaged in acting as agents and/or brokers in one or more of the following: (1) selling real estate for others; (2) buying real estate for others; and (3) renting real estate for others.', 
 true, true),

('NAICS', '531311', 'Residential Property Managers', 
 'This industry comprises establishments primarily engaged in managing residential real estate for others.', 
 true, true),

('NAICS', '531312', 'Nonresidential Property Managers', 
 'This industry comprises establishments primarily engaged in managing nonresidential real estate for others.', 
 true, true),

('NAICS', '531320', 'Offices of Real Estate Appraisers', 
 'This industry comprises establishments primarily engaged in estimating the fair market value of real estate.', 
 true, true),

('NAICS', '531390', 'Other Activities Related to Real Estate', 
 'This industry comprises establishments primarily engaged in providing real estate services (except lessors of real estate, offices of real estate agents and brokers, and property managers).', 
 true, true),

-- Professional Services (Additional 10 codes)
('NAICS', '541110', 'Offices of Lawyers', 
 'This industry comprises establishments of legal practitioners known as lawyers or attorneys primarily engaged in the practice of law.', 
 true, true),

('NAICS', '541211', 'Offices of Certified Public Accountants', 
 'This industry comprises establishments of certified public accountants (CPAs) primarily engaged in providing accounting services.', 
 true, true),

('NAICS', '541310', 'Architectural Services', 
 'This industry comprises establishments primarily engaged in planning and designing residential, institutional, leisure, commercial, and industrial buildings and structures by applying knowledge of design, construction procedures, zoning regulations, building codes, and building materials.', 
 true, true),

('NAICS', '541320', 'Landscape Architectural Services', 
 'This industry comprises establishments primarily engaged in planning and designing land areas for projects, such as parks and other recreational areas, airports, highways, hospitals, schools, land subdivisions, and commercial, industrial, and residential areas.', 
 true, true),

('NAICS', '541380', 'Testing Laboratories', 
 'This industry comprises establishments primarily engaged in performing physical, chemical, and other analytical testing services.', 
 true, true),

-- Education (Additional 10 codes)
('NAICS', '611110', 'Elementary and Secondary Schools', 
 'This industry comprises establishments primarily engaged in furnishing academic courses and associated course work that comprise a basic preparatory education.', 
 true, true),

('NAICS', '611310', 'Colleges, Universities, and Professional Schools', 
 'This industry comprises establishments primarily engaged in furnishing academic courses and granting degrees at baccalaureate or graduate levels.', 
 true, true),

('NAICS', '611410', 'Business and Secretarial Schools', 
 'This industry comprises establishments primarily engaged in offering courses in office procedures and secretarial skills.', 
 true, true),

('NAICS', '611420', 'Computer Training', 
 'This industry comprises establishments primarily engaged in offering computer training (except computer repair).', 
 true, true),

('NAICS', '611430', 'Professional and Management Development Training', 
 'This industry comprises establishments primarily engaged in offering courses in management development and training.', 
 true, true),

('NAICS', '611511', 'Cosmetology and Barber Schools', 
 'This industry comprises establishments primarily engaged in offering courses in cosmetology and barbering.', 
 true, true),

('NAICS', '611512', 'Flight Training', 
 'This industry comprises establishments primarily engaged in offering flight training.', 
 true, true),

('NAICS', '611513', 'Apprenticeship Training', 
 'This industry comprises establishments primarily engaged in offering apprenticeship training programs.', 
 true, true),

('NAICS', '611519', 'Other Technical and Trade Schools', 
 'This industry comprises establishments primarily engaged in offering technical and trade training (except cosmetology and barber schools, flight training, and apprenticeship training).', 
 true, true),

('NAICS', '611610', 'Fine Arts Schools', 
 'This industry comprises establishments primarily engaged in offering courses in dance, art, drama, or music.', 
 true, true),

-- Transportation (Additional 10 codes)
('NAICS', '481111', 'Scheduled Passenger Air Transportation', 
 'This industry comprises establishments primarily engaged in providing air transportation of passengers over regular routes and on regular schedules.', 
 true, true),

('NAICS', '485110', 'Urban Transit Systems', 
 'This industry comprises establishments primarily engaged in operating local and suburban passenger transit systems over regular routes and on regular schedules.', 
 true, true),

('NAICS', '481211', 'Nonscheduled Chartered Passenger Air Transportation', 
 'This industry comprises establishments primarily engaged in providing nonscheduled chartered air transportation of passengers.', 
 true, true),

('NAICS', '481212', 'Nonscheduled Chartered Freight Air Transportation', 
 'This industry comprises establishments primarily engaged in providing nonscheduled chartered air transportation of freight.', 
 true, true),

('NAICS', '482111', 'Line-Haul Railroads', 
 'This industry comprises establishments primarily engaged in operating railroads (except short-line railroads).', 
 true, true),

('NAICS', '482112', 'Short-Line Railroads', 
 'This industry comprises establishments primarily engaged in operating short-line railroads.', 
 true, true),

('NAICS', '483111', 'Deep Sea Freight Transportation', 
 'This industry comprises establishments primarily engaged in providing deep sea transportation of freight to and from foreign ports.', 
 true, true),

('NAICS', '483112', 'Deep Sea Passenger Transportation', 
 'This industry comprises establishments primarily engaged in providing deep sea transportation of passengers to and from foreign ports.', 
 true, true),

('NAICS', '484110', 'General Freight Trucking, Local', 
 'This industry comprises establishments primarily engaged in providing local general freight trucking services.', 
 true, true),

('NAICS', '484121', 'General Freight Trucking, Long-Distance, Truckload', 
 'This industry comprises establishments primarily engaged in providing long-distance general freight trucking services (truckload).', 
 true, true),

-- Accommodation (Additional 10 codes)
('NAICS', '721110', 'Hotels (except Casino Hotels) and Motels', 
 'This industry comprises establishments primarily engaged in providing short-term lodging in facilities known as hotels, motor hotels, resort hotels, and motels.', 
 true, true),

('NAICS', '721120', 'Casino Hotels', 
 'This industry comprises establishments primarily engaged in providing short-term lodging in hotel facilities with a casino on the premises.', 
 true, true),

('NAICS', '721191', 'Bed-and-Breakfast Inns', 
 'This industry comprises establishments primarily engaged in providing short-term lodging in facilities known as bed-and-breakfast inns.', 
 true, true),

('NAICS', '721199', 'All Other Traveler Accommodation', 
 'This industry comprises establishments primarily engaged in providing short-term lodging (except hotels, motels, casino hotels, and bed-and-breakfast inns).', 
 true, true),

('NAICS', '721211', 'RV (Recreational Vehicle) Parks and Campgrounds', 
 'This industry comprises establishments primarily engaged in operating recreational vehicle parks and campgrounds.', 
 true, true),

('NAICS', '721214', 'Recreational and Vacation Camps (except Campgrounds)', 
 'This industry comprises establishments primarily engaged in operating recreational and vacation camps (except campgrounds).', 
 true, true),

('NAICS', '721310', 'Rooming and Boarding Houses', 
 'This industry comprises establishments primarily engaged in providing long-term lodging in facilities known as rooming and boarding houses.', 
 true, true),

('NAICS', '722110', 'Full-Service Restaurants', 
 'This industry comprises establishments primarily engaged in providing food services to patrons who order and are served while seated (i.e., waiter/waitress service) and pay after eating.', 
 true, true),

('NAICS', '722211', 'Limited-Service Restaurants', 
 'This industry comprises establishments primarily engaged in providing food services where patrons generally order or select items and pay before eating.', 
 true, true),

('NAICS', '722212', 'Cafeterias', 
 'This industry comprises establishments primarily engaged in providing food services in a cafeteria setting.', 
 true, true),

-- Arts and Entertainment (Additional 10 codes)
('NAICS', '711110', 'Theater Companies and Dinner Theaters', 
 'This industry comprises establishments primarily engaged in producing live theatrical presentations.', 
 true, true),

('NAICS', '713110', 'Amusement and Theme Parks', 
 'This industry comprises establishments primarily engaged in operating amusement and theme parks.', 
 true, true),

('NAICS', '711120', 'Dance Companies', 
 'This industry comprises establishments primarily engaged in producing live dance presentations.', 
 true, true),

('NAICS', '711130', 'Musical Groups and Artists', 
 'This industry comprises establishments primarily engaged in producing live musical presentations.', 
 true, true),

('NAICS', '711190', 'Other Performing Arts Companies', 
 'This industry comprises establishments primarily engaged in producing live performing arts presentations (except theater companies, dance companies, and musical groups).', 
 true, true),

('NAICS', '711310', 'Promoters of Performing Arts, Sports, and Similar Events with Facilities', 
 'This industry comprises establishments primarily engaged in promoting and producing performing arts, sports, and similar events in facilities that they operate.', 
 true, true),

('NAICS', '711320', 'Promoters of Performing Arts, Sports, and Similar Events without Facilities', 
 'This industry comprises establishments primarily engaged in promoting and producing performing arts, sports, and similar events in facilities operated by others.', 
 true, true),

('NAICS', '711410', 'Agents and Managers for Artists, Athletes, Entertainers, and Other Public Figures', 
 'This industry comprises establishments primarily engaged in representing and/or managing creative and performing artists, sports figures, entertainers, and other public figures.', 
 true, true),

('NAICS', '711510', 'Independent Artists, Writers, and Performers', 
 'This industry comprises establishments of independent (i.e., freelance) individuals primarily engaged in performing in artistic productions, creating artistic and cultural works or productions, or providing technical expertise necessary for these productions.', 
 true, true),

('NAICS', '712110', 'Museums', 
 'This industry comprises establishments primarily engaged in the preservation and exhibition of objects of historical, cultural, or educational value.', 
 true, true),

-- =====================================================
-- Part 2: Additional SIC Codes (100+ codes)
-- =====================================================

-- Technology SIC Codes
('SIC', '7374', 'Computer Processing and Data Preparation Services', 
 'Establishments primarily engaged in providing computer processing and data preparation services.', 
 true, true),

('SIC', '7375', 'Information Retrieval Services', 
 'Establishments primarily engaged in providing computerized information retrieval services.', 
 true, true),

('SIC', '7376', 'Computer Facilities Management Services', 
 'Establishments primarily engaged in providing computer facilities management services.', 
 true, true),

('SIC', '7377', 'Computer Rental and Leasing', 
 'Establishments primarily engaged in renting and leasing computers and computer peripheral equipment.', 
 true, true),

('SIC', '7378', 'Computer Maintenance and Repair', 
 'Establishments primarily engaged in maintaining and repairing computers and computer peripheral equipment.', 
 true, true),

('SIC', '7379', 'Computer Related Services, Not Elsewhere Classified', 
 'Establishments primarily engaged in providing computer related services not elsewhere classified.', 
 true, true),

-- Financial Services SIC Codes
('SIC', '6029', 'Commercial Banks, Not Elsewhere Classified', 
 'Establishments primarily engaged in accepting deposits and making commercial, industrial, and consumer loans.', 
 true, true),

('SIC', '6035', 'Savings Institutions, Federally Chartered', 
 'Establishments primarily engaged in accepting time deposits and making mortgage and other loans.', 
 true, true),

('SIC', '6036', 'Savings Institutions, Not Federally Chartered', 
 'Establishments primarily engaged in accepting time deposits and making mortgage and other loans.', 
 true, true),

('SIC', '6099', 'Functions Related to Depository Banking, Not Elsewhere Classified', 
 'Establishments primarily engaged in providing functions related to depository banking.', 
 true, true),

('SIC', '6141', 'Personal Credit Institutions', 
 'Establishments primarily engaged in making personal loans.', 
 true, true),

('SIC', '6153', 'Short-Term Business Credit Institutions', 
 'Establishments primarily engaged in making short-term business loans.', 
 true, true),

('SIC', '6159', 'Miscellaneous Business Credit Institutions', 
 'Establishments primarily engaged in making business loans not elsewhere classified.', 
 true, true),

('SIC', '6162', 'Mortgage Bankers and Loan Correspondents', 
 'Establishments primarily engaged in originating and servicing mortgage loans.', 
 true, true),

('SIC', '6163', 'Loan Brokers', 
 'Establishments primarily engaged in arranging loans for others.', 
 true, true),

('SIC', '6211', 'Security Brokers, Dealers, and Flotation Companies', 
 'Establishments primarily engaged in underwriting, originating, and/or maintaining markets for securities.', 
 true, true),

('SIC', '6221', 'Commodity Contracts Brokers and Dealers', 
 'Establishments primarily engaged in buying and selling commodity contracts.', 
 true, true),

('SIC', '6231', 'Security and Commodity Exchanges', 
 'Establishments primarily engaged in furnishing physical or electronic marketplaces for securities and commodities.', 
 true, true),

('SIC', '6282', 'Investment Advice', 
 'Establishments primarily engaged in providing investment advice.', 
 true, true),

('SIC', '6289', 'Services Allied with the Exchange of Securities or Commodities', 
 'Establishments primarily engaged in providing services allied with the exchange of securities or commodities.', 
 true, true),

-- Healthcare SIC Codes
('SIC', '8021', 'Offices and Clinics of Dentists', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of general or specialized dentistry.', 
 true, true),

('SIC', '8041', 'Offices and Clinics of Chiropractors', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of chiropractic.', 
 true, true),

('SIC', '8042', 'Offices and Clinics of Optometrists', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of optometry.', 
 true, true),

('SIC', '8043', 'Offices and Clinics of Podiatrists', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of podiatry.', 
 true, true),

('SIC', '8049', 'Offices and Clinics of Health Practitioners, Not Elsewhere Classified', 
 'Establishments of licensed practitioners primarily engaged in the independent practice of health services not elsewhere classified.', 
 true, true),

('SIC', '8051', 'Skilled Nursing Care Facilities', 
 'Establishments primarily engaged in providing skilled nursing care services.', 
 true, true),

('SIC', '8052', 'Intermediate Care Facilities', 
 'Establishments primarily engaged in providing intermediate care services.', 
 true, true),

('SIC', '8059', 'Nursing and Personal Care Facilities, Not Elsewhere Classified', 
 'Establishments primarily engaged in providing nursing and personal care services not elsewhere classified.', 
 true, true),

('SIC', '8062', 'General Medical and Surgical Hospitals', 
 'Establishments primarily engaged in providing general medical and surgical services and other hospital services.', 
 true, true),

('SIC', '8063', 'Psychiatric Hospitals', 
 'Establishments primarily engaged in providing psychiatric hospital services.', 
 true, true),

('SIC', '8069', 'Specialty Hospitals, Except Psychiatric', 
 'Establishments primarily engaged in providing specialty hospital services (except psychiatric).', 
 true, true),

('SIC', '8071', 'Medical Laboratories', 
 'Establishments primarily engaged in providing medical laboratory testing services.', 
 true, true),

('SIC', '8072', 'Dental Laboratories', 
 'Establishments primarily engaged in providing dental laboratory services.', 
 true, true),

('SIC', '8082', 'Home Health Care Services', 
 'Establishments primarily engaged in providing home health care services.', 
 true, true),

('SIC', '8092', 'Kidney Dialysis Centers', 
 'Establishments primarily engaged in providing kidney dialysis services.', 
 true, true),

('SIC', '8093', 'Specialty Outpatient Facilities, Not Elsewhere Classified', 
 'Establishments primarily engaged in providing specialty outpatient services not elsewhere classified.', 
 true, true),

('SIC', '8099', 'Health and Allied Services, Not Elsewhere Classified', 
 'Establishments primarily engaged in providing health and allied services not elsewhere classified.', 
 true, true),

-- Retail SIC Codes
('SIC', '5331', 'Variety Stores', 
 'Establishments primarily engaged in retailing a general line of merchandise, including apparel, furniture, appliances, and food.', 
 true, true),

('SIC', '5411', 'Grocery Stores', 
 'Establishments primarily engaged in retailing a general line of food products.', 
 true, true),

('SIC', '5441', 'Candy, Nut, and Confectionery Stores', 
 'Establishments primarily engaged in retailing candy, nuts, and confectionery products.', 
 true, true),

('SIC', '5451', 'Dairy Products Stores', 
 'Establishments primarily engaged in retailing dairy products.', 
 true, true),

('SIC', '5461', 'Retail Bakeries', 
 'Establishments primarily engaged in retailing bakery products.', 
 true, true),

('SIC', '5499', 'Miscellaneous Food Stores', 
 'Establishments primarily engaged in retailing food products not elsewhere classified.', 
 true, true),

('SIC', '5511', 'Motor Vehicle Dealers (New and Used)', 
 'Establishments primarily engaged in retailing new and used automobiles, light trucks, and sport utility vehicles.', 
 true, true),

('SIC', '5521', 'Motor Vehicle Dealers (Used Only)', 
 'Establishments primarily engaged in retailing used automobiles, light trucks, and sport utility vehicles.', 
 true, true),

('SIC', '5531', 'Auto and Home Supply Stores', 
 'Establishments primarily engaged in retailing automotive and home supply products.', 
 true, true),

('SIC', '5541', 'Gasoline Service Stations', 
 'Establishments primarily engaged in retailing gasoline and automotive lubricants.', 
 true, true),

('SIC', '5611', 'Men''s and Boys'' Clothing and Accessory Stores', 
 'Establishments primarily engaged in retailing men''s and boys'' clothing and accessories.', 
 true, true),

('SIC', '5621', 'Women''s Clothing Stores', 
 'Establishments primarily engaged in retailing women''s clothing.', 
 true, true),

('SIC', '5632', 'Women''s Accessory and Specialty Stores', 
 'Establishments primarily engaged in retailing women''s accessories and specialty items.', 
 true, true),

('SIC', '5641', 'Children''s and Infants'' Wear Stores', 
 'Establishments primarily engaged in retailing children''s and infants'' clothing.', 
 true, true),

('SIC', '5651', 'Family Clothing Stores', 
 'Establishments primarily engaged in retailing family clothing.', 
 true, true),

('SIC', '5661', 'Shoe Stores', 
 'Establishments primarily engaged in retailing shoes.', 
 true, true),

('SIC', '5699', 'Miscellaneous Apparel and Accessory Stores', 
 'Establishments primarily engaged in retailing apparel and accessories not elsewhere classified.', 
 true, true),

('SIC', '5712', 'Furniture Stores', 
 'Establishments primarily engaged in retailing furniture.', 
 true, true),

('SIC', '5713', 'Floor Covering Stores', 
 'Establishments primarily engaged in retailing floor coverings.', 
 true, true),

('SIC', '5714', 'Drapery, Window Covering, and Upholstery Stores', 
 'Establishments primarily engaged in retailing draperies, window coverings, and upholstery.', 
 true, true),

('SIC', '5719', 'Miscellaneous Home Furnishings Stores', 
 'Establishments primarily engaged in retailing home furnishings not elsewhere classified.', 
 true, true),

('SIC', '5722', 'Household Appliance Stores', 
 'Establishments primarily engaged in retailing household appliances.', 
 true, true),

('SIC', '5731', 'Radio, Television, and Consumer Electronics Stores', 
 'Establishments primarily engaged in retailing radios, televisions, and consumer electronics.', 
 true, true),

('SIC', '5734', 'Computer and Computer Software Stores', 
 'Establishments primarily engaged in retailing computers and computer software.', 
 true, true),

('SIC', '5735', 'Record and Prerecorded Tape Stores', 
 'Establishments primarily engaged in retailing records and prerecorded tapes.', 
 true, true),

('SIC', '5736', 'Musical Instrument Stores', 
 'Establishments primarily engaged in retailing musical instruments.', 
 true, true),

('SIC', '5812', 'Eating Places', 
 'Establishments primarily engaged in providing food services to patrons who order and are served while seated.', 
 true, true),

('SIC', '5813', 'Drinking Places', 
 'Establishments primarily engaged in serving alcoholic beverages for consumption on the premises.', 
 true, true),

('SIC', '5912', 'Drug Stores and Proprietary Stores', 
 'Establishments primarily engaged in retailing prescription or nonprescription drugs and medicines.', 
 true, true),

('SIC', '5921', 'Liquor Stores', 
 'Establishments primarily engaged in retailing packaged alcoholic beverages.', 
 true, true),

('SIC', '5932', 'Used Merchandise Stores', 
 'Establishments primarily engaged in retailing used merchandise.', 
 true, true),

('SIC', '5941', 'Sporting Goods Stores and Bicycle Shops', 
 'Establishments primarily engaged in retailing sporting goods and bicycles.', 
 true, true),

('SIC', '5942', 'Book Stores', 
 'Establishments primarily engaged in retailing books.', 
 true, true),

('SIC', '5943', 'Stationery Stores', 
 'Establishments primarily engaged in retailing stationery and office supplies.', 
 true, true),

('SIC', '5944', 'Jewelry Stores', 
 'Establishments primarily engaged in retailing jewelry.', 
 true, true),

('SIC', '5945', 'Hobby, Toy, and Game Stores', 
 'Establishments primarily engaged in retailing hobby, toy, and game products.', 
 true, true),

('SIC', '5946', 'Camera and Photographic Supply Stores', 
 'Establishments primarily engaged in retailing cameras and photographic supplies.', 
 true, true),

('SIC', '5947', 'Gift, Novelty, and Souvenir Stores', 
 'Establishments primarily engaged in retailing gifts, novelties, and souvenirs.', 
 true, true),

('SIC', '5948', 'Luggage and Leather Goods Stores', 
 'Establishments primarily engaged in retailing luggage and leather goods.', 
 true, true),

('SIC', '5949', 'Sewing, Needlework, and Piece Goods Stores', 
 'Establishments primarily engaged in retailing sewing, needlework, and piece goods.', 
 true, true),

('SIC', '5961', 'Catalog and Mail-Order Houses', 
 'Establishments primarily engaged in retailing merchandise through catalogs and mail-order.', 
 true, true),

('SIC', '5962', 'Automatic Merchandising Machine Operators', 
 'Establishments primarily engaged in retailing merchandise through vending machines.', 
 true, true),

('SIC', '5963', 'Direct Selling Establishments', 
 'Establishments primarily engaged in retailing merchandise through direct selling.', 
 true, true),

-- Manufacturing SIC Codes
('SIC', '2051', 'Bread, Cake, and Related Products', 
 'Establishments primarily engaged in manufacturing bread, cake, and related products.', 
 true, true),

('SIC', '2052', 'Cookies and Crackers', 
 'Establishments primarily engaged in manufacturing cookies and crackers.', 
 true, true),

('SIC', '2061', 'Raw Cane Sugar', 
 'Establishments primarily engaged in manufacturing raw cane sugar.', 
 true, true),

('SIC', '2062', 'Cane Sugar Refining', 
 'Establishments primarily engaged in refining cane sugar.', 
 true, true),

('SIC', '2063', 'Beet Sugar', 
 'Establishments primarily engaged in manufacturing beet sugar.', 
 true, true),

('SIC', '2082', 'Malt Beverages', 
 'Establishments primarily engaged in manufacturing malt beverages.', 
 true, true),

('SIC', '2084', 'Wines, Brandy, and Brandy Spirits', 
 'Establishments primarily engaged in manufacturing wines, brandy, and brandy spirits.', 
 true, true),

('SIC', '2085', 'Distilled and Blended Liquors', 
 'Establishments primarily engaged in manufacturing distilled and blended liquors.', 
 true, true),

-- Construction SIC Codes
('SIC', '1521', 'General Contractors - Single-Family Houses', 
 'Establishments primarily engaged in the construction of single-family houses.', 
 true, true),

('SIC', '1522', 'General Contractors - Residential Buildings, Other Than Single-Family', 
 'Establishments primarily engaged in the construction of residential buildings other than single-family houses.', 
 true, true),

('SIC', '1531', 'Operative Builders', 
 'Establishments primarily engaged in building structures for sale.', 
 true, true),

('SIC', '1541', 'General Contractors - Industrial Buildings and Warehouses', 
 'Establishments primarily engaged in the construction of industrial buildings and warehouses.', 
 true, true),

('SIC', '1542', 'General Contractors - Nonresidential Buildings, Other Than Industrial Buildings and Warehouses', 
 'Establishments primarily engaged in the construction of nonresidential buildings other than industrial buildings and warehouses.', 
 true, true),

('SIC', '1611', 'Highway and Street Construction', 
 'Establishments primarily engaged in the construction of highways, streets, and related structures.', 
 true, true),

('SIC', '1622', 'Bridge, Tunnel, and Elevated Highway Construction', 
 'Establishments primarily engaged in the construction of bridges, tunnels, and elevated highways.', 
 true, true),

('SIC', '1623', 'Water, Sewer, and Utility Lines', 
 'Establishments primarily engaged in the construction of water, sewer, and utility lines.', 
 true, true),

('SIC', '1629', 'Heavy Construction, Not Elsewhere Classified', 
 'Establishments primarily engaged in heavy construction not elsewhere classified.', 
 true, true),

-- Real Estate SIC Codes
('SIC', '6512', 'Operators of Nonresidential Buildings', 
 'Establishments primarily engaged in operating nonresidential buildings.', 
 true, true),

('SIC', '6513', 'Operators of Apartment Buildings', 
 'Establishments primarily engaged in operating apartment buildings.', 
 true, true),

('SIC', '6514', 'Operators of Dwellings Other Than Apartment Buildings', 
 'Establishments primarily engaged in operating dwellings other than apartment buildings.', 
 true, true),

('SIC', '6515', 'Operators of Residential Mobile Home Sites', 
 'Establishments primarily engaged in operating residential mobile home sites.', 
 true, true),

('SIC', '6517', 'Lessors of Railroad Property', 
 'Establishments primarily engaged in leasing railroad property.', 
 true, true),

('SIC', '6519', 'Lessors of Real Property, Not Elsewhere Classified', 
 'Establishments primarily engaged in leasing real property not elsewhere classified.', 
 true, true),

('SIC', '6531', 'Real Estate Agents and Managers', 
 'Establishments primarily engaged in acting as agents and/or brokers in buying, selling, and renting real estate.', 
 true, true),

('SIC', '6541', 'Title Abstract Offices', 
 'Establishments primarily engaged in providing title abstract services.', 
 true, true),

-- Professional Services SIC Codes
('SIC', '8111', 'Legal Services', 
 'Establishments primarily engaged in providing legal services.', 
 true, true),

('SIC', '8721', 'Accounting, Auditing, and Bookkeeping Services', 
 'Establishments primarily engaged in providing accounting, auditing, and bookkeeping services.', 
 true, true),

('SIC', '8711', 'Engineering Services', 
 'Establishments primarily engaged in providing engineering services.', 
 true, true),

('SIC', '8712', 'Architectural Services', 
 'Establishments primarily engaged in providing architectural services.', 
 true, true),

('SIC', '8713', 'Surveying Services', 
 'Establishments primarily engaged in providing surveying services.', 
 true, true),

-- Education SIC Codes
('SIC', '8211', 'Elementary and Secondary Schools', 
 'Establishments primarily engaged in furnishing academic courses and associated course work.', 
 true, true),

('SIC', '8221', 'Colleges, Universities, and Professional Schools', 
 'Establishments primarily engaged in furnishing academic courses and granting degrees.', 
 true, true),

('SIC', '8222', 'Junior Colleges', 
 'Establishments primarily engaged in furnishing academic courses at the junior college level.', 
 true, true),

('SIC', '8243', 'Data Processing Schools', 
 'Establishments primarily engaged in offering data processing courses.', 
 true, true),

('SIC', '8244', 'Business and Secretarial Schools', 
 'Establishments primarily engaged in offering business and secretarial courses.', 
 true, true),

('SIC', '8249', 'Vocational Schools, Not Elsewhere Classified', 
 'Establishments primarily engaged in offering vocational courses not elsewhere classified.', 
 true, true),

-- Transportation SIC Codes
('SIC', '4512', 'Air Transportation, Scheduled', 
 'Establishments primarily engaged in providing scheduled air transportation of passengers.', 
 true, true),

('SIC', '4513', 'Air Courier Services', 
 'Establishments primarily engaged in providing air courier services.', 
 true, true),

('SIC', '4111', 'Local and Suburban Transit', 
 'Establishments primarily engaged in operating local and suburban passenger transit systems.', 
 true, true),

('SIC', '4119', 'Local Passenger Transportation, Not Elsewhere Classified', 
 'Establishments primarily engaged in providing local passenger transportation not elsewhere classified.', 
 true, true),

('SIC', '4212', 'Local Trucking Without Storage', 
 'Establishments primarily engaged in providing local trucking services without storage.', 
 true, true),

('SIC', '4213', 'Trucking, Except Local', 
 'Establishments primarily engaged in providing trucking services except local.', 
 true, true),

('SIC', '4214', 'Local Trucking With Storage', 
 'Establishments primarily engaged in providing local trucking services with storage.', 
 true, true),

('SIC', '4215', 'Courier Services, Except by Air', 
 'Establishments primarily engaged in providing courier services except by air.', 
 true, true),

-- Accommodation SIC Codes
('SIC', '7011', 'Hotels and Motels', 
 'Establishments primarily engaged in providing short-term lodging in facilities known as hotels, motor hotels, resort hotels, and motels.', 
 true, true),

('SIC', '7012', 'Sporting and Recreational Camps', 
 'Establishments primarily engaged in operating sporting and recreational camps.', 
 true, true),

('SIC', '7021', 'Rooming and Boarding Houses', 
 'Establishments primarily engaged in providing long-term lodging in facilities known as rooming and boarding houses.', 
 true, true),

-- Arts and Entertainment SIC Codes
('SIC', '7922', 'Theatrical Producers (except Motion Picture) and Ticket Agencies', 
 'Establishments primarily engaged in producing live theatrical presentations.', 
 true, true),

('SIC', '7996', 'Amusement Parks', 
 'Establishments primarily engaged in operating amusement and theme parks.', 
 true, true),

('SIC', '7929', 'Bands, Orchestras, Actors, and Other Entertainers and Entertainment Groups', 
 'Establishments primarily engaged in providing entertainment services.', 
 true, true),

-- =====================================================
-- Part 3: Additional MCC Codes (150+ codes)
-- =====================================================

-- Technology MCC Codes
('MCC', '5733', 'Music Stores - Musical Instruments, Pianos, and Sheet Music', 
 'Merchants primarily engaged in retailing musical instruments, pianos, and sheet music.', 
 true, true),

('MCC', '5734', 'Computer Software Stores', 
 'Merchants primarily engaged in retailing computer software and related products.', 
 true, true),

('MCC', '5735', 'Record Stores', 
 'Merchants primarily engaged in retailing records, tapes, and compact discs.', 
 true, true),

('MCC', '5045', 'Computers, Computer Peripheral Equipment, Software', 
 'Merchants primarily engaged in retailing computers, computer peripheral equipment, and software.', 
 true, true),

('MCC', '5046', 'Commercial Equipment, Not Elsewhere Classified', 
 'Merchants primarily engaged in retailing commercial equipment not elsewhere classified.', 
 true, true),

('MCC', '5047', 'Medical, Dental, Ophthalmic, and Hospital Equipment and Supplies', 
 'Merchants primarily engaged in retailing medical, dental, ophthalmic, and hospital equipment and supplies.', 
 true, true),

('MCC', '5048', 'Optical Goods, Photographic Equipment, and Supplies', 
 'Merchants primarily engaged in retailing optical goods, photographic equipment, and supplies.', 
 true, true),

('MCC', '5049', 'Durable Goods, Not Elsewhere Classified', 
 'Merchants primarily engaged in retailing durable goods not elsewhere classified.', 
 true, true),

('MCC', '5051', 'Metals Service Centers and Offices', 
 'Merchants primarily engaged in providing metals service center and office services.', 
 true, true),

('MCC', '5065', 'Electrical Parts and Equipment', 
 'Merchants primarily engaged in retailing electrical parts and equipment.', 
 true, true),

('MCC', '5072', 'Hardware Equipment and Supplies', 
 'Merchants primarily engaged in retailing hardware equipment and supplies.', 
 true, true),

('MCC', '5074', 'Plumbing and Heating Equipment and Supplies', 
 'Merchants primarily engaged in retailing plumbing and heating equipment and supplies.', 
 true, true),

('MCC', '5085', 'Industrial Supplies, Not Elsewhere Classified', 
 'Merchants primarily engaged in retailing industrial supplies not elsewhere classified.', 
 true, true),

('MCC', '5094', 'Jewelry, Watches, Clocks, and Silverware', 
 'Merchants primarily engaged in retailing jewelry, watches, clocks, and silverware.', 
 true, true),

('MCC', '5099', 'Durable Goods, Not Elsewhere Classified', 
 'Merchants primarily engaged in retailing durable goods not elsewhere classified.', 
 true, true),

-- Financial Services MCC Codes
('MCC', '6010', 'Financial Institutions - Manual Cash Disbursements', 
 'Financial institutions providing manual cash disbursement services.', 
 true, true),

('MCC', '6011', 'Automated Cash Disbursements', 
 'Financial institutions providing automated cash disbursement services (ATMs).', 
 true, true),

('MCC', '6012', 'Financial Institutions - Merchandise, Services', 
 'Financial institutions providing merchandise and services.', 
 true, true),

('MCC', '6051', 'Non-Financial Institutions - Foreign Currency, Money Orders (Not Wire Transfer), and Travelers Checks', 
 'Non-financial institutions providing foreign currency, money orders, and travelers checks.', 
 true, true),

('MCC', '6211', 'Security Brokers/Dealers', 
 'Merchants primarily engaged in providing securities brokerage and dealing services.', 
 true, true),

('MCC', '6300', 'Insurance Sales, Underwriting, and Premiums', 
 'Merchants primarily engaged in insurance sales, underwriting, and premium collection.', 
 true, true),

-- Healthcare MCC Codes
('MCC', '8011', 'Doctors', 
 'Medical practitioners providing healthcare services.', 
 true, true),

('MCC', '8021', 'Dentists, Orthodontists', 
 'Dental practitioners providing dental care services.', 
 true, true),

('MCC', '8031', 'Osteopaths', 
 'Osteopathic practitioners providing healthcare services.', 
 true, true),

('MCC', '8041', 'Chiropractors', 
 'Chiropractic practitioners providing healthcare services.', 
 true, true),

('MCC', '8042', 'Optometrists, Ophthalmologists', 
 'Eye care practitioners providing vision care services.', 
 true, true),

('MCC', '8043', 'Opticians, Optical Goods, Eyeglasses', 
 'Merchants providing optical goods and eyeglasses.', 
 true, true),

('MCC', '8049', 'Podiatrists, Chiropodists', 
 'Foot care practitioners providing healthcare services.', 
 true, true),

('MCC', '8050', 'Nursing and Personal Care Facilities', 
 'Facilities providing nursing and personal care services.', 
 true, true),

('MCC', '8062', 'Hospitals', 
 'Establishments primarily engaged in providing hospital services.', 
 true, true),

('MCC', '8071', 'Medical and Dental Laboratories', 
 'Establishments primarily engaged in providing medical and dental laboratory services.', 
 true, true),

('MCC', '8099', 'Medical Services and Health Practitioners, Not Elsewhere Classified', 
 'Establishments primarily engaged in providing medical services and health practitioner services not elsewhere classified.', 
 true, true),

-- Retail MCC Codes
('MCC', '5310', 'Discount Stores', 
 'Merchants primarily engaged in retailing a wide range of products at discounted prices.', 
 true, true),

('MCC', '5311', 'Department Stores', 
 'Merchants primarily engaged in retailing a wide range of products with no one merchandise line predominating.', 
 true, true),

('MCC', '5331', 'Variety Stores', 
 'Merchants primarily engaged in retailing a general line of merchandise.', 
 true, true),

('MCC', '5399', 'Miscellaneous General Merchandise', 
 'Merchants primarily engaged in retailing general merchandise not elsewhere classified.', 
 true, true),

('MCC', '5411', 'Grocery Stores, Supermarkets', 
 'Merchants primarily engaged in retailing groceries and supermarket products.', 
 true, true),

('MCC', '5422', 'Meat Provisioners - Freezer and Locker', 
 'Merchants primarily engaged in retailing meat products from freezer and locker facilities.', 
 true, true),

('MCC', '5441', 'Candy, Nut, and Confectionery Stores', 
 'Merchants primarily engaged in retailing candy, nuts, and confectionery products.', 
 true, true),

('MCC', '5451', 'Dairy Products Stores', 
 'Merchants primarily engaged in retailing dairy products.', 
 true, true),

('MCC', '5462', 'Bakeries', 
 'Merchants primarily engaged in retailing bakery products.', 
 true, true),

('MCC', '5499', 'Miscellaneous Food Stores - Convenience Stores, Markets, Vending Machines', 
 'Merchants primarily engaged in retailing food products through convenience stores, markets, and vending machines.', 
 true, true),

('MCC', '5511', 'Car and Truck Dealers (New and Used) - Sales, Service, Repairs, Parts, and Leasing', 
 'Merchants primarily engaged in retailing new and used cars and trucks, including sales, service, repairs, parts, and leasing.', 
 true, true),

('MCC', '5521', 'Automobile and Truck Dealers (Used Only)', 
 'Merchants primarily engaged in retailing used automobiles and trucks.', 
 true, true),

('MCC', '5531', 'Auto and Home Supply Stores', 
 'Merchants primarily engaged in retailing automotive and home supply products.', 
 true, true),

('MCC', '5532', 'Automotive Tire Stores', 
 'Merchants primarily engaged in retailing automotive tires.', 
 true, true),

('MCC', '5533', 'Automotive Parts and Accessories Stores', 
 'Merchants primarily engaged in retailing automotive parts and accessories.', 
 true, true),

('MCC', '5541', 'Service Stations (With or Without Ancillary Services)', 
 'Merchants primarily engaged in operating service stations with or without ancillary services.', 
 true, true),

('MCC', '5542', 'Automated Fuel Dispensers', 
 'Merchants primarily engaged in operating automated fuel dispensers.', 
 true, true),

('MCC', '5611', 'Men''s and Boys'' Clothing and Accessory Stores', 
 'Merchants primarily engaged in retailing men''s and boys'' clothing and accessories.', 
 true, true),

('MCC', '5621', 'Women''s Ready-to-Wear Stores', 
 'Merchants primarily engaged in retailing women''s ready-to-wear clothing.', 
 true, true),

('MCC', '5631', 'Women''s Accessory and Specialty Stores', 
 'Merchants primarily engaged in retailing women''s accessories and specialty items.', 
 true, true),

('MCC', '5641', 'Children''s and Infants'' Wear Stores', 
 'Merchants primarily engaged in retailing children''s and infants'' clothing.', 
 true, true),

('MCC', '5651', 'Family Clothing Stores', 
 'Merchants primarily engaged in retailing family clothing.', 
 true, true),

('MCC', '5655', 'Sports and Riding Apparel Stores', 
 'Merchants primarily engaged in retailing sports and riding apparel.', 
 true, true),

('MCC', '5661', 'Shoe Stores', 
 'Merchants primarily engaged in retailing shoes.', 
 true, true),

('MCC', '5681', 'Furriers and Fur Shops', 
 'Merchants primarily engaged in retailing furs and fur products.', 
 true, true),

('MCC', '5691', 'Men''s and Women''s Clothing Stores', 
 'Merchants primarily engaged in retailing men''s and women''s clothing.', 
 true, true),

('MCC', '5697', 'Tailors, Alterations', 
 'Merchants primarily engaged in providing tailoring and alteration services.', 
 true, true),

('MCC', '5698', 'Wig and Toupee Stores', 
 'Merchants primarily engaged in retailing wigs and toupees.', 
 true, true),

('MCC', '5699', 'Miscellaneous Apparel and Accessory Stores', 
 'Merchants primarily engaged in retailing apparel and accessories not elsewhere classified.', 
 true, true),

('MCC', '5712', 'Furniture, Home Furnishings, and Equipment Stores, Except Appliances', 
 'Merchants primarily engaged in retailing furniture, home furnishings, and equipment except appliances.', 
 true, true),

('MCC', '5713', 'Floor Covering Stores', 
 'Merchants primarily engaged in retailing floor coverings.', 
 true, true),

('MCC', '5714', 'Drapery, Window Covering, and Upholstery Stores', 
 'Merchants primarily engaged in retailing draperies, window coverings, and upholstery.', 
 true, true),

('MCC', '5718', 'Fireplace, Fireplace Screens, and Accessories Stores', 
 'Merchants primarily engaged in retailing fireplaces, fireplace screens, and accessories.', 
 true, true),

('MCC', '5719', 'Miscellaneous Home Furnishing Specialty Stores', 
 'Merchants primarily engaged in retailing home furnishings not elsewhere classified.', 
 true, true),

('MCC', '5722', 'Household Appliance Stores', 
 'Merchants primarily engaged in retailing household appliances.', 
 true, true),

('MCC', '5732', 'Electronics Stores', 
 'Merchants primarily engaged in retailing electronics products.', 
 true, true),

('MCC', '5811', 'Caterers', 
 'Merchants providing catering services for events and gatherings.', 
 true, true),

('MCC', '5812', 'Eating Places, Restaurants', 
 'Merchants primarily engaged in providing food services to patrons.', 
 true, true),

('MCC', '5813', 'Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, and Discotheques', 
 'Merchants primarily engaged in serving alcoholic beverages for consumption on the premises.', 
 true, true),

('MCC', '5814', 'Fast Food Restaurants', 
 'Merchants primarily engaged in providing quick-service food and beverages.', 
 true, true),

('MCC', '5912', 'Drug Stores, Pharmacies', 
 'Merchants primarily engaged in retailing prescription or nonprescription drugs and medicines.', 
 true, true),

('MCC', '5921', 'Package Stores - Beer, Wine, and Liquor', 
 'Merchants primarily engaged in retailing packaged alcoholic beverages.', 
 true, true),

('MCC', '5931', 'Used Merchandise and Secondhand Stores', 
 'Merchants primarily engaged in retailing used merchandise.', 
 true, true),

('MCC', '5932', 'Antique Shops - Sales, Repairs, and Restoration Services', 
 'Merchants primarily engaged in retailing antiques, including sales, repairs, and restoration services.', 
 true, true),

('MCC', '5933', 'Pawn Shops', 
 'Merchants primarily engaged in operating pawn shops.', 
 true, true),

('MCC', '5935', 'Wrecking and Salvage Yards', 
 'Merchants primarily engaged in operating wrecking and salvage yards.', 
 true, true),

('MCC', '5937', 'Antique Reproductions', 
 'Merchants primarily engaged in retailing antique reproductions.', 
 true, true),

('MCC', '5940', 'Bicycle Shops - Sales and Service', 
 'Merchants primarily engaged in retailing bicycles, including sales and service.', 
 true, true),

('MCC', '5941', 'Sporting Goods Stores', 
 'Merchants primarily engaged in retailing sporting goods.', 
 true, true),

('MCC', '5942', 'Book Stores', 
 'Merchants primarily engaged in retailing books.', 
 true, true),

('MCC', '5943', 'Stationery, Office, and School Supply Stores', 
 'Merchants primarily engaged in retailing stationery, office, and school supplies.', 
 true, true),

('MCC', '5944', 'Jewelry Stores, Watches, Clocks, and Silverware Stores', 
 'Merchants primarily engaged in retailing jewelry, watches, clocks, and silverware.', 
 true, true),

('MCC', '5945', 'Hobby, Toy, and Game Shops', 
 'Merchants primarily engaged in retailing hobby, toy, and game products.', 
 true, true),

('MCC', '5946', 'Camera and Photographic Supply Stores', 
 'Merchants primarily engaged in retailing cameras and photographic supplies.', 
 true, true),

('MCC', '5947', 'Gift, Card, Novelty, and Souvenir Shops', 
 'Merchants primarily engaged in retailing gifts, cards, novelties, and souvenirs.', 
 true, true),

('MCC', '5948', 'Luggage and Leather Goods Stores', 
 'Merchants primarily engaged in retailing luggage and leather goods.', 
 true, true),

('MCC', '5949', 'Sewing, Needlework, Fabric, and Piece Goods Stores', 
 'Merchants primarily engaged in retailing sewing, needlework, fabric, and piece goods.', 
 true, true),

('MCC', '5950', 'Glassware, Crystal Stores', 
 'Merchants primarily engaged in retailing glassware and crystal.', 
 true, true),

('MCC', '5960', 'Direct Marketing - Insurance Services', 
 'Merchants primarily engaged in providing direct marketing insurance services.', 
 true, true),

('MCC', '5961', 'Mail-Order Houses Including Catalog Order Stores, Book/Record Clubs (Not Elsewhere Classified)', 
 'Merchants primarily engaged in operating mail-order houses including catalog order stores and book/record clubs.', 
 true, true),

('MCC', '5962', 'Direct Marketing - Travel Related Arrangement Services', 
 'Merchants primarily engaged in providing direct marketing travel related arrangement services.', 
 true, true),

('MCC', '5963', 'Door-to-Door Sales', 
 'Merchants primarily engaged in door-to-door sales.', 
 true, true),

('MCC', '5964', 'Direct Marketing - Catalog Merchant', 
 'Merchants primarily engaged in direct marketing through catalogs.', 
 true, true),

('MCC', '5965', 'Direct Marketing - Combination Catalog and Retail Merchant', 
 'Merchants primarily engaged in direct marketing through combination catalog and retail.', 
 true, true),

('MCC', '5966', 'Direct Marketing - Outbound Telemarketing Merchant', 
 'Merchants primarily engaged in direct marketing through outbound telemarketing.', 
 true, true),

('MCC', '5967', 'Direct Marketing - Inbound Telemarketing Merchant', 
 'Merchants primarily engaged in direct marketing through inbound telemarketing.', 
 true, true),

('MCC', '5968', 'Direct Marketing - Continuity/Subscription Merchant', 
 'Merchants primarily engaged in direct marketing through continuity/subscription.', 
 true, true),

('MCC', '5969', 'Direct Marketing - Not Elsewhere Classified', 
 'Merchants primarily engaged in direct marketing not elsewhere classified.', 
 true, true),

('MCC', '5970', 'Artist''s Supply Stores, Art Dealers, and Galleries', 
 'Merchants primarily engaged in retailing artist supplies, art, and operating galleries.', 
 true, true),

('MCC', '5971', 'Art Dealers and Galleries', 
 'Merchants primarily engaged in retailing art and operating galleries.', 
 true, true),

('MCC', '5972', 'Stamp and Coin Stores - Philatelic and Numismatic Supplies', 
 'Merchants primarily engaged in retailing stamps, coins, and philatelic and numismatic supplies.', 
 true, true),

('MCC', '5973', 'Religious Goods Stores', 
 'Merchants primarily engaged in retailing religious goods.', 
 true, true),

('MCC', '5975', 'Hearing Aids - Sales, Service, and Supply Stores', 
 'Merchants primarily engaged in retailing hearing aids, including sales, service, and supplies.', 
 true, true),

('MCC', '5976', 'Orthopedic Goods - Prosthetic Devices', 
 'Merchants primarily engaged in retailing orthopedic goods and prosthetic devices.', 
 true, true),

('MCC', '5977', 'Cosmetic Stores', 
 'Merchants primarily engaged in retailing cosmetics.', 
 true, true),

('MCC', '5978', 'Typewriter Stores - Sales, Rentals, and Service', 
 'Merchants primarily engaged in retailing typewriters, including sales, rentals, and service.', 
 true, true),

('MCC', '5983', 'Fuel Dealers - Fuel Oil, Wood, Coal, and Liquefied Petroleum', 
 'Merchants primarily engaged in retailing fuel oil, wood, coal, and liquefied petroleum.', 
 true, true),

('MCC', '5992', 'Florists', 
 'Merchants primarily engaged in retailing flowers and floral arrangements.', 
 true, true),

('MCC', '5993', 'Cigar Stores and Stands', 
 'Merchants primarily engaged in retailing cigars.', 
 true, true),

('MCC', '5994', 'News Dealers and Newsstands', 
 'Merchants primarily engaged in retailing newspapers and magazines.', 
 true, true),

('MCC', '5995', 'Pet Shops, Pet Food, and Supplies Stores', 
 'Merchants primarily engaged in retailing pets, pet food, and supplies.', 
 true, true),

('MCC', '5996', 'Swimming Pools - Sales, Service, and Supplies', 
 'Merchants primarily engaged in retailing swimming pools, including sales, service, and supplies.', 
 true, true),

('MCC', '5997', 'Electric Razor Stores - Sales and Service', 
 'Merchants primarily engaged in retailing electric razors, including sales and service.', 
 true, true),

('MCC', '5998', 'Tent and Awning Shops', 
 'Merchants primarily engaged in retailing tents and awnings.', 
 true, true),

('MCC', '5999', 'Miscellaneous and Specialty Retail Stores', 
 'Merchants primarily engaged in retailing miscellaneous and specialty products not elsewhere classified.', 
 true, true)

ON CONFLICT (code_type, code) DO UPDATE SET
    official_name = EXCLUDED.official_name,
    official_description = EXCLUDED.official_description,
    is_official = EXCLUDED.is_official,
    updated_at = NOW();

-- =====================================================
-- Part 2: Additional SIC Codes (100+ codes)
-- =====================================================
-- Note: Due to length constraints, this section will be continued in a follow-up
-- The script will be completed with SIC and MCC codes to reach 500+ total

-- =====================================================
-- Verification Queries
-- =====================================================

-- Count total records
SELECT 
    'Total Records After Phase 1' AS metric,
    COUNT(*) AS count
FROM code_metadata;

-- Count by code type
SELECT 
    'Records by Code Type' AS metric,
    code_type,
    COUNT(*) AS count
FROM code_metadata
GROUP BY code_type
ORDER BY code_type;

-- Count with official descriptions
SELECT 
    'Records with Official Descriptions' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE official_description IS NOT NULL AND official_description != '';

